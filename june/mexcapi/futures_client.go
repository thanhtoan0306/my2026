package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type FuturesClient struct {
	APIKey string
	Secret string
	HTTP   *http.Client
}

type FuturesPosition struct {
	PositionID     int64   `json:"positionId"`
	Symbol         string  `json:"symbol"`
	PositionType   int     `json:"positionType"`
	OpenType       int     `json:"openType"`
	State          int     `json:"state"`
	HoldVol        float64 `json:"holdVol"`
	HoldAvgPrice   float64 `json:"holdAvgPrice"`
	LiquidatePrice float64 `json:"liquidatePrice"`
	IM             float64 `json:"im"`
	Realised       float64 `json:"realised"`
	Leverage       int     `json:"leverage"`
	ProfitRatio    float64 `json:"profitRatio"`
	CreateTime     int64   `json:"createTime"`
	UpdateTime     int64   `json:"updateTime"`
}

type PositionView struct {
	Position      FuturesPosition
	DisplaySymbol string
	SideLabel     string
	SideClass     string
	MarginMode    string
	FairPrice     float64
	UnrealizedPNL float64
	PNLPercent    float64
	HasPNL        bool
}

type futuresEnvelope struct {
	Success bool            `json:"success"`
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func NewFuturesClient(apiKey, secret string) *FuturesClient {
	return &FuturesClient{
		APIKey: apiKey,
		Secret: secret,
		HTTP: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *FuturesClient) OpenPositions() ([]FuturesPosition, error) {
	body, err := c.signedGet("/api/v1/private/position/open_positions", nil)
	if err != nil {
		return nil, err
	}
	var positions []FuturesPosition
	if err := json.Unmarshal(body, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}

func (c *FuturesClient) signedGet(path string, params map[string]string) (json.RawMessage, error) {
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)

	paramStr := ""
	if len(params) > 0 {
		values := url.Values{}
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			values.Set(k, params[k])
		}
		paramStr = values.Encode()
	}

	signTarget := c.APIKey + ts + paramStr
	sig := futuresSign(signTarget, c.Secret)

	reqURL := mexcBase + path
	if paramStr != "" {
		reqURL += "?" + paramStr
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("ApiKey", c.APIKey)
	req.Header.Set("Request-Time", ts)
	req.Header.Set("Signature", sig)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var env futuresEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("parse futures response: %w", err)
	}
	if !env.Success || env.Code != 0 {
		msg := env.Message
		if msg == "" {
			msg = strings.TrimSpace(string(raw))
		}
		return nil, fmt.Errorf("%s (code %d)", msg, env.Code)
	}
	return env.Data, nil
}

func futuresSign(target, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(target))
	return hex.EncodeToString(mac.Sum(nil))
}

func (c *FuturesClient) loadPositionViews() ([]PositionView, error) {
	positions, err := c.OpenPositions()
	if err != nil {
		return nil, err
	}

	fairPrices, err := fetchFairPrices(c.HTTP)
	if err != nil {
		return nil, err
	}
	contractSizes, err := fetchContractSizes(c.HTTP)
	if err != nil {
		return nil, err
	}

	views := make([]PositionView, 0, len(positions))
	for _, p := range positions {
		if p.State != 1 {
			continue
		}
		v := buildPositionView(p, fairPrices[p.Symbol], contractSizes[p.Symbol])
		views = append(views, v)
	}

	sort.Slice(views, func(i, j int) bool {
		if views[i].UnrealizedPNL == views[j].UnrealizedPNL {
			return views[i].Position.Symbol < views[j].Position.Symbol
		}
		return views[i].UnrealizedPNL > views[j].UnrealizedPNL
	})

	return views, nil
}

func buildPositionView(p FuturesPosition, fairPrice, contractSize float64) PositionView {
	v := PositionView{Position: p}
	v.DisplaySymbol = strings.ReplaceAll(p.Symbol, "_", "/") + " Perpetual"

	switch p.PositionType {
	case 1:
		v.SideLabel = fmt.Sprintf("%dX Long", p.Leverage)
		v.SideClass = "buy"
	case 2:
		v.SideLabel = fmt.Sprintf("%dX Short", p.Leverage)
		v.SideClass = "sell"
	default:
		v.SideLabel = fmt.Sprintf("%dX", p.Leverage)
	}

	switch p.OpenType {
	case 1:
		v.MarginMode = "Isolated"
	case 2:
		v.MarginMode = "Cross"
	default:
		v.MarginMode = "—"
	}

	if fairPrice > 0 && contractSize > 0 && p.HoldVol > 0 {
		v.FairPrice = fairPrice
		v.HasPNL = true
		size := p.HoldVol * contractSize
		if p.PositionType == 1 {
			v.UnrealizedPNL = (fairPrice - p.HoldAvgPrice) * size
		} else {
			v.UnrealizedPNL = (p.HoldAvgPrice - fairPrice) * size
		}
		if p.IM > 0 {
			v.PNLPercent = (v.UnrealizedPNL / p.IM) * 100
		}
	}

	return v
}

func fetchFairPrices(client *http.Client) (map[string]float64, error) {
	resp, err := client.Get(mexcBase + "/api/v1/contract/ticker")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var env struct {
		Success bool `json:"success"`
		Data    []struct {
			Symbol    string  `json:"symbol"`
			FairPrice float64 `json:"fairPrice"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}

	out := make(map[string]float64, len(env.Data))
	for _, t := range env.Data {
		out[t.Symbol] = t.FairPrice
	}
	return out, nil
}

func fetchContractSizes(client *http.Client) (map[string]float64, error) {
	resp, err := client.Get(mexcBase + "/api/v1/contract/detail")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var env struct {
		Success bool `json:"success"`
		Data    []struct {
			Symbol       string  `json:"symbol"`
			ContractSize float64 `json:"contractSize"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}

	out := make(map[string]float64, len(env.Data))
	for _, d := range env.Data {
		out[d.Symbol] = d.ContractSize
	}
	return out, nil
}
