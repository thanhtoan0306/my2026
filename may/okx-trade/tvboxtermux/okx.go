package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const okxRESTBase = "https://www.okx.com"

type OKXClient struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	Demo       bool
	HTTP       *http.Client
}

type okxResponse struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type Position struct {
	InstID   string `json:"instId"`
	Lever    string `json:"lever"`
	Upl      string `json:"upl"`
	UplRatio string `json:"uplRatio"`
	AvgPx    string `json:"avgPx"`
	MarkPx   string `json:"markPx"`
	IdxPx    string `json:"idxPx"`
	Pos      string `json:"pos"`
	Margin   string `json:"margin"`
	MgnMode  string `json:"mgnMode"`
}

func (p Position) DisplayName() string {
	name := p.InstID
	name = replaceSuffix(name, "-SWAP", "")
	name = replaceAll(name, "-", "")
	return name + " Perpetual"
}

func (p Position) UplFloat() float64   { return parseFloat(p.Upl) }
func (p Position) UplRatioPct() string {
	r := parseFloat(p.UplRatio) * 100
	sign := ""
	if r >= 0 {
		sign = "+"
	}
	return fmt.Sprintf("%s%.2f", sign, r)
}
func (p Position) IsPositive() bool { return p.UplFloat() >= 0 }
func (p Position) PnlSign() string {
	if p.UplFloat() >= 0 {
		return "+"
	}
	return ""
}
func (p Position) BreakevenPx() string {
	px := p.IdxPx
	if px == "" {
		px = p.AvgPx
	}
	return formatPrice(px)
}
func (p Position) MarkPrice() string  { return formatPrice(p.MarkPx) }
func (p Position) EntryPrice() string { return formatPrice(p.AvgPx) }
func (p Position) SizeUSDT() string {
	return fmt.Sprintf("%.2f USDT", parseFloat(p.Pos))
}
func (p Position) MarginUSDT() string {
	return fmt.Sprintf("%.2f USDT", parseFloat(p.Margin))
}
func (p Position) UplUSDT() string {
	return fmt.Sprintf("%s%.2f USDT", p.PnlSign(), p.UplFloat())
}

func NewOKXClient(apiKey, secret, passphrase string, demo bool) *OKXClient {
	return &OKXClient{
		APIKey:     apiKey,
		SecretKey:  secret,
		Passphrase: passphrase,
		Demo:       demo,
		HTTP:       &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *OKXClient) Sign(timestamp, method, path, body string) string {
	return c.sign(timestamp, method, path, body)
}

func (c *OKXClient) sign(timestamp, method, path, body string) string {
	msg := timestamp + method + path + body
	mac := hmac.New(sha256.New, []byte(c.SecretKey))
	mac.Write([]byte(msg))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (c *OKXClient) setAuthHeaders(req *http.Request, timestamp, method, path, body string) {
	req.Header.Set("OK-ACCESS-KEY", c.APIKey)
	req.Header.Set("OK-ACCESS-SIGN", c.sign(timestamp, method, path, body))
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", c.Passphrase)
	req.Header.Set("Content-Type", "application/json")
	if c.Demo {
		req.Header.Set("x-simulated-trading", "1")
		req.Header.Set("x-simulated-auth", "1")
	}
}

func (c *OKXClient) GetPositions() ([]Position, error) {
	const path = "/api/v5/account/positions"
	ts := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	req, err := http.NewRequest(http.MethodGet, okxRESTBase+path, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeaders(req, ts, http.MethodGet, path, "")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result okxResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Code != "0" {
		return nil, fmt.Errorf("OKX API: %s (code %s)", result.Msg, result.Code)
	}

	var positions []Position
	if err := json.Unmarshal(result.Data, &positions); err != nil {
		return nil, err
	}

	open := make([]Position, 0, len(positions))
	for _, p := range positions {
		if parseFloat(p.Pos) != 0 {
			open = append(open, p)
		}
	}
	return open, nil
}

func (c *OKXClient) ClosePosition(instID, mgnMode string) error {
	const path = "/api/v5/trade/close-position"
	ts := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	payload := map[string]string{
		"instId":  instID,
		"mgnMode": mgnMode,
	}
	raw, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, okxRESTBase+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	c.setAuthHeaders(req, ts, http.MethodPost, path, string(raw))

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result okxResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	if result.Code != "0" {
		return fmt.Errorf("close position: %s (code %s)", result.Msg, result.Code)
	}
	return nil
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func formatPrice(s string) string {
	return fmt.Sprintf("₮%.5f", parseFloat(s))
}

func replaceSuffix(s, old, new string) string {
	if len(s) >= len(old) && s[len(s)-len(old):] == old {
		return s[:len(s)-len(old)] + new
	}
	return s
}

func replaceAll(s, old, new string) string {
	out := ""
	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			out += new
			i += len(old)
		} else {
			out += string(s[i])
			i++
		}
	}
	return out
}
