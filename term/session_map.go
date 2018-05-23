package term

import (
	"sync"
	"time"
)

type SessionMap struct {
	sync.Mutex
	Sessions map[string]*SocketIoTermSession
}

func (m *SessionMap) Set(token string, session *SocketIoTermSession) {
	m.Lock()
	defer m.Unlock()
	m.Sessions[token] = session
}

func (m *SessionMap) Get(token string) (*SocketIoTermSession, bool) {
	m.Lock()
	defer m.Unlock()
	session, ok := m.Sessions[token]
	return session, ok
}

func (m *SessionMap) CleanUp(ttl time.Duration) {
	m.Lock()
	defer m.Unlock()

	now := time.Now()
	expiredfrom := now.Add(-ttl)

	tokens := []string{}
	for token, session := range m.Sessions {
		if session.Detached && session.CreatedAt.Before(expiredfrom) {
			session.Terminate()
			tokens = append(tokens, token)
		}
	}

	for _, token := range tokens {
		delete(m.Sessions, token)
	}
}
