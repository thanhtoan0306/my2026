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

const mexcBase = "https://api.mexc.com"

type Client struct {
	APIKey string
	Secret string
	HTTP   *http.Client
}

type Account struct {
	CanTrade    bool      `json:"canTrade"`
	CanWithdraw bool      `json:"canWithdraw"`
	CanDeposit  bool      `json:"canDeposit"`
	AccountType string    `json:"accountType"`
	Balances    []Balance `json:"balances"`
}

type Balance struct {
	Asset     string `json:"asset"`
	Free      string `json:"free"`
	Locked    string `json:"locked"`
	Available string `json:"available"`
}

type Trade struct {
	Symbol          string `json:"symbol"`
	ID              string `json:"id"`
	OrderID         string `json:"orderId"`
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	QuoteQty        string `json:"quoteQty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	Time            int64  `json:"time"`
	IsBuyer         bool   `json:"isBuyer"`
	IsMaker         bool   `json:"isMaker"`
}

type Order struct {
	Symbol      string `json:"symbol"`
	OrderID     string `json:"orderId"`
	Price       string `json:"price"`
	OrigQty     string `json:"origQty"`
	ExecutedQty string `json:"executedQty"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Side        string `json:"side"`
	Time        int64  `json:"time"`
}

type apiError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewClient(apiKey, secret string) *Client {
	return &Client{
		APIKey: apiKey,
		Secret: secret,
		HTTP: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) ServerTime() (int64, error) {
	body, err := c.publicGet("/api/v3/time")
	if err != nil {
		return 0, err
	}
	var out struct {
		ServerTime int64 `json:"serverTime"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, err
	}
	return out.ServerTime, nil
}

func (c *Client) Account() (*Account, error) {
	body, err := c.signedGet("/api/v3/account", nil)
	if err != nil {
		return nil, err
	}
	var acct Account
	if err := json.Unmarshal(body, &acct); err != nil {
		return nil, err
	}
	return &acct, nil
}

func (c *Client) MyTrades(symbol string, limit int) ([]Trade, error) {
	body, err := c.signedGet("/api/v3/myTrades", map[string]string{
		"symbol": symbol,
		"limit":  strconv.Itoa(limit),
	})
	if err != nil {
		return nil, err
	}
	var trades []Trade
	if err := json.Unmarshal(body, &trades); err != nil {
		return nil, err
	}
	return trades, nil
}

func (c *Client) AllOrders(symbol string, limit int) ([]Order, error) {
	body, err := c.signedGet("/api/v3/allOrders", map[string]string{
		"symbol": symbol,
		"limit":  strconv.Itoa(limit),
	})
	if err != nil {
		return nil, err
	}
	var orders []Order
	if err := json.Unmarshal(body, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (c *Client) publicGet(path string) ([]byte, error) {
	resp, err := c.HTTP.Get(mexcBase + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return readAPI(resp)
}

func (c *Client) signedGet(path string, params map[string]string) ([]byte, error) {
	ts, err := c.ServerTime()
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	values.Set("recvWindow", "5000")
	values.Set("timestamp", strconv.FormatInt(ts, 10))

	query := values.Encode()
	sig := sign(query, c.Secret)
	fullURL := mexcBase + path + "?" + query + "&signature=" + sig

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-MEXC-APIKEY", c.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return readAPI(resp)
}

func readAPI(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Msg != "" {
			return nil, fmt.Errorf("%s (%d)", apiErr.Msg, apiErr.Code)
		}
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var apiErr apiError
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Code != 0 {
		return nil, fmt.Errorf("%s (%d)", apiErr.Msg, apiErr.Code)
	}
	return body, nil
}

func sign(query, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(query))
	return hex.EncodeToString(mac.Sum(nil))
}

var quoteAssets = []string{"USDT", "USDC", "BTC", "ETH"}

func symbolsFromBalances(balances []Balance, current string) []string {
	seen := map[string]bool{current: true}
	var out []string
	if current != "" {
		out = append(out, current)
	}

	for _, b := range balances {
		free, _ := strconv.ParseFloat(b.Free, 64)
		locked, _ := strconv.ParseFloat(b.Locked, 64)
		if free+locked <= 0 {
			continue
		}
		isQuote := false
		for _, q := range quoteAssets {
			if b.Asset == q {
				isQuote = true
				break
			}
		}
		if isQuote {
			continue
		}
		for _, q := range quoteAssets {
			sym := b.Asset + q
			if !seen[sym] {
				seen[sym] = true
				out = append(out, sym)
			}
		}
	}

	sort.Strings(out[1:])
	return out
}

func nonZeroBalances(balances []Balance) []Balance {
	var out []Balance
	for _, b := range balances {
		free, _ := strconv.ParseFloat(b.Free, 64)
		locked, _ := strconv.ParseFloat(b.Locked, 64)
		if free+locked > 0 {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		fi, _ := strconv.ParseFloat(out[i].Free, 64)
		fj, _ := strconv.ParseFloat(out[j].Free, 64)
		return fi > fj
	})
	return out
}
