package main

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Session struct {
	mu sync.RWMutex

	APIKey     string
	SecretKey  string
	Passphrase string
	Demo       bool
	Positions  []Position
	Connected  bool
	StatusText string
	Logs       []string

	wsRunning bool
	wsStop    chan struct{}
}

type SessionStore struct {
	mu   sync.RWMutex
	byID map[string]*Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{byID: make(map[string]*Session)}
}

func (s *SessionStore) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.byID[id]
	return sess, ok
}

func (s *SessionStore) Set(id string, sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[id] = sess
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.byID[id]; ok {
		sess.stopWS()
	}
	delete(s.byID, id)
}

func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (sess *Session) addLog(msg string) {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	const maxLogs = 50
	sess.Logs = append([]string{msg}, sess.Logs...)
	if len(sess.Logs) > maxLogs {
		sess.Logs = sess.Logs[:maxLogs]
	}
}

func (sess *Session) client() *OKXClient {
	return NewOKXClient(sess.APIKey, sess.SecretKey, sess.Passphrase, sess.Demo)
}

func (sess *Session) refreshPositions() error {
	positions, err := sess.client().GetPositions()
	if err != nil {
		return err
	}
	sess.setPositions(positions)
	sess.mu.Lock()
	sess.Connected = true
	sess.mu.Unlock()
	return nil
}

func (sess *Session) setPositions(positions []Position) {
	open := filterOpenPositions(positions)
	sess.mu.Lock()
	sess.Positions = open
	sess.mu.Unlock()
}

func filterOpenPositions(positions []Position) []Position {
	open := make([]Position, 0, len(positions))
	for _, p := range positions {
		if parseFloat(p.Pos) != 0 {
			open = append(open, p)
		}
	}
	return open
}

func (sess *Session) setConnected(status string) {
	sess.mu.Lock()
	sess.Connected = true
	sess.StatusText = status
	sess.mu.Unlock()
}

func (sess *Session) setDisconnected(status string) {
	sess.mu.Lock()
	sess.Connected = false
	sess.StatusText = status
	sess.mu.Unlock()
}

func (sess *Session) snapshot() (positions []Position, connected bool, status string, logs []string, demo bool) {
	sess.mu.RLock()
	defer sess.mu.RUnlock()
	positions = append([]Position(nil), sess.Positions...)
	logs = append([]string(nil), sess.Logs...)
	return positions, sess.Connected, sess.StatusText, logs, sess.Demo
}

func (sess *Session) startWS(sessionID string) {
	sess.mu.Lock()
	if sess.wsRunning {
		sess.mu.Unlock()
		return
	}
	sess.wsRunning = true
	sess.wsStop = make(chan struct{})
	stop := sess.wsStop
	sess.mu.Unlock()
	go runOKXPrivateWS(sessionID, sess, stop)
}

func (sess *Session) stopWS() {
	sess.mu.Lock()
	if !sess.wsRunning {
		sess.mu.Unlock()
		return
	}
	close(sess.wsStop)
	sess.wsRunning = false
	sess.mu.Unlock()
}
