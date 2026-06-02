package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/contrib/auth/terminal"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

const defaultChatID = int64(-5045031903)

type ChatMessage struct {
	ID        int    `json:"id"`
	Date      string `json:"date"`
	From      string `json:"from,omitempty"`
	Text      string `json:"text,omitempty"`
	MediaType string `json:"mediaType,omitempty"`
}

type SyncPayload struct {
	ChatID   int64         `json:"chatId"`
	ChatURL  string        `json:"chatUrl"`
	Date     string        `json:"date"`
	SyncedAt string        `json:"syncedAt"`
	Count    int           `json:"count"`
	Messages []ChatMessage `json:"messages"`
}

func chatIDFromEnv() int64 {
	if v := os.Getenv("TELEGRAM_CHAT_ID"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return n
		}
	}
	return defaultChatID
}

func apiCredentials() (int, string, error) {
	idStr := os.Getenv("TELEGRAM_API_ID")
	hash := os.Getenv("TELEGRAM_API_HASH")
	if idStr == "" || hash == "" {
		return 0, "", fmt.Errorf("set TELEGRAM_API_ID and TELEGRAM_API_HASH from https://my.telegram.org/apps")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, "", fmt.Errorf("invalid TELEGRAM_API_ID: %w", err)
	}
	return id, hash, nil
}

func sessionDir() string {
	if d := os.Getenv("TELEGRAM_SESSION_DIR"); d != "" {
		return d
	}
	return "session"
}

func peerForChatID(chatID int64) tg.InputPeerClass {
	if chatID > 0 {
		return &tg.InputPeerUser{UserID: chatID}
	}
	abs := -chatID
	if abs > 1_000_000_000_000 {
		channelID := abs - 1_000_000_000_000
		return &tg.InputPeerChannel{ChannelID: channelID}
	}
	return &tg.InputPeerChat{ChatID: abs}
}

func messageText(m *tg.Message) string {
	if m == nil {
		return ""
	}
	return m.Message
}

func mediaType(m *tg.Message) string {
	if m == nil || m.Media == nil {
		return ""
	}
	switch m.Media.(type) {
	case *tg.MessageMediaPhoto:
		return "photo"
	case *tg.MessageMediaDocument:
		return "document"
	case *tg.MessageMediaWebPage:
		return "webpage"
	default:
		return fmt.Sprintf("%T", m.Media)
	}
}

func normalizeHistory(hist tg.MessagesMessagesClass) (*tg.MessagesMessages, error) {
	switch h := hist.(type) {
	case *tg.MessagesMessages:
		return h, nil
	case *tg.MessagesMessagesSlice:
		return &tg.MessagesMessages{Messages: h.Messages, Users: h.Users, Chats: h.Chats}, nil
	case *tg.MessagesChannelMessages:
		return &tg.MessagesMessages{Messages: h.Messages, Users: h.Users, Chats: h.Chats}, nil
	default:
		return nil, fmt.Errorf("unexpected history type %T", hist)
	}
}

func syncTodayMessages(ctx context.Context) (*SyncPayload, error) {
	apiID, apiHash, err := apiCredentials()
	if err != nil {
		return nil, err
	}

	chatID := chatIDFromEnv()
	loc := time.Local
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	payload := &SyncPayload{
		ChatID:   chatID,
		ChatURL:  fmt.Sprintf("https://web.telegram.org/k/#%d", chatID),
		Date:     startOfDay.Format("2006-01-02"),
		SyncedAt: now.Format(time.RFC3339),
		Messages: []ChatMessage{},
	}

	userNames := map[int64]string{}

	client := telegram.NewClient(apiID, apiHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: sessionDir()},
	})

	err = client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, auth.NewFlow(
			terminal.New(os.Stdin, os.Stdout),
			auth.SendCodeOptions{},
		)); err != nil {
			return err
		}

		api := client.API()
		peer := peerForChatID(chatID)
		offsetID := 0

		for {
			hist, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer:     peer,
				Limit:    100,
				OffsetID: offsetID,
			})
			if err != nil {
				return fmt.Errorf("get history: %w", err)
			}

			box, err := normalizeHistory(hist)
			if err != nil {
				return err
			}

			for _, u := range box.Users {
				if user, ok := u.(*tg.User); ok {
					label := strings.TrimSpace(user.FirstName + " " + user.LastName)
					if user.Username != "" {
						label = "@" + user.Username
					}
					if label == "" {
						label = strconv.FormatInt(user.ID, 10)
					}
					userNames[user.ID] = label
				}
			}

			if len(box.Messages) == 0 {
				break
			}

			stop := false
			minID := 0

			for _, item := range box.Messages {
				msg, ok := item.(*tg.Message)
				if !ok {
					continue
				}
				t := time.Unix(int64(msg.Date), 0).In(loc)
				if minID == 0 || msg.ID < minID {
					minID = msg.ID
				}
				if t.Before(startOfDay) {
					stop = true
					continue
				}
				from := ""
				if msg.FromID != nil {
					if f, ok := msg.FromID.(*tg.PeerUser); ok {
						from = userNames[f.UserID]
					}
				}
				payload.Messages = append(payload.Messages, ChatMessage{
					ID:        msg.ID,
					Date:      t.Format(time.RFC3339),
					From:      from,
					Text:      messageText(msg),
					MediaType: mediaType(msg),
				})
			}

			if stop || minID == 0 {
				break
			}
			offsetID = minID
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	payload.Count = len(payload.Messages)
	log.Printf("synced %d messages for chat %d on %s", payload.Count, chatID, payload.Date)
	return payload, nil
}

func marshalPayload(p *SyncPayload) ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

func runLogin(ctx context.Context) error {
	apiID, apiHash, err := apiCredentials()
	if err != nil {
		return err
	}
	client := telegram.NewClient(apiID, apiHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: sessionDir()},
	})
	return client.Run(ctx, func(ctx context.Context) error {
		return client.Auth().IfNecessary(ctx, auth.NewFlow(
			terminal.New(os.Stdin, os.Stdout),
			auth.SendCodeOptions{},
		))
	})
}
