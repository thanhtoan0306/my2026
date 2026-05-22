package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (
	okxWSReal = "wss://ws.okx.com:443/ws/v5/private"
	okxWSDemo = "wss://wspap.okx.com:443/ws/v5/private"
)

type wsEnvelope struct {
	Event string          `json:"event"`
	Code  string          `json:"code"`
	Msg   string          `json:"msg"`
	Arg   *wsArg          `json:"arg"`
	Data  json.RawMessage `json:"data"`
}

type wsArg struct {
	Channel string `json:"channel"`
}

func okxWSURL(demo bool) string {
	if demo {
		return okxWSDemo
	}
	return okxWSReal
}

func runOKXPrivateWS(sessionID string, sess *Session, stop <-chan struct{}) {
	backoff := 3 * time.Second
	for {
		select {
		case <-stop:
			return
		default:
		}
		err := okxWSConnectOnce(sess, stop)
		if err != nil {
			sess.addLog("WS: " + err.Error())
			sess.setDisconnected("WS mất kết nối")
		}
		select {
		case <-stop:
			return
		case <-time.After(backoff):
		}
	}
}

func okxWSConnectOnce(sess *Session, stop <-chan struct{}) error {
	client := sess.client()
	sess.addLog("WS: đang kết nối " + okxWSURL(sess.Demo) + "...")

	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, _, err := dialer.Dial(okxWSURL(sess.Demo), nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		select {
		case <-stop:
			_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
			conn.Close()
		case <-done:
		}
	}()
	defer close(done)

	if err := okxWSLogin(conn, client); err != nil {
		return err
	}
	sess.addLog("WS: login OK, subscribe positions SWAP")
	sess.setConnected("Đã kết nối · WS live")

	sub := map[string]any{
		"op": "subscribe",
		"args": []map[string]string{{
			"channel":  "positions",
			"instType": "SWAP",
		}},
	}
	if err := conn.WriteJSON(sub); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	for {
		select {
		case <-stop:
			return nil
		default:
		}
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		var msg wsEnvelope
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		if msg.Event == "error" {
			return fmt.Errorf("ws error: %s (code %s)", msg.Msg, msg.Code)
		}
		if msg.Event == "login" && msg.Code != "0" {
			return fmt.Errorf("login failed: %s (code %s)", msg.Msg, msg.Code)
		}
		if msg.Arg != nil && msg.Arg.Channel == "positions" && len(msg.Data) > 0 {
			var positions []Position
			if err := json.Unmarshal(msg.Data, &positions); err != nil {
				continue
			}
			sess.setPositions(positions)
		}
	}
}

func okxWSLogin(conn *websocket.Conn, client *OKXClient) error {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sign := client.Sign(ts, "GET", "/users/self/verify", "")
	login := map[string]any{
		"op": "login",
		"args": []map[string]string{{
			"apiKey":     client.APIKey,
			"passphrase": client.Passphrase,
			"timestamp":  ts,
			"sign":       sign,
		}},
	}
	if err := conn.WriteJSON(login); err != nil {
		return fmt.Errorf("login send: %w", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("login read: %w", err)
	}
	var msg wsEnvelope
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}
	if msg.Event == "login" && msg.Code == "0" {
		return nil
	}
	return fmt.Errorf("login: %s (code %s)", msg.Msg, msg.Code)
}
