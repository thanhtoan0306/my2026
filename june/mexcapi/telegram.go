package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Telegram struct {
	Token  string
	ChatID string
	HTTP   *http.Client
}

func NewTelegram(token, chatID string) *Telegram {
	return &Telegram{
		Token:  token,
		ChatID: chatID,
		HTTP: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (t *Telegram) Enabled() bool {
	return t != nil && t.Token != "" && t.ChatID != ""
}

func (t *Telegram) Send(text string) error {
	if !t.Enabled() {
		return fmt.Errorf("telegram not configured (set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID in .env)")
	}

	body, err := json.Marshal(map[string]string{
		"chat_id": t.ChatID,
		"text":    text,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.Token)
	resp, err := t.HTTP.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram api: %s", string(raw))
	}

	var out struct {
		OK bool `json:"ok"`
	}
	if json.Unmarshal(raw, &out) == nil && !out.OK {
		return fmt.Errorf("telegram api: %s", string(raw))
	}
	return nil
}
