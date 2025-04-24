package states

import (
	"sync"
)

type Manager struct {
	mu     sync.Mutex
	states map[int64]State // userID -> states
}

func NewManager() *Manager {
	return &Manager{
		states: make(map[int64]State),
	}
}

func (m *Manager) Get(userID int64) State {
	m.mu.Lock()
	defer m.mu.Unlock()

	if state, ok := m.states[userID]; ok {
		return state
	}
	return StateDefault // Важно возвращать StateDefault, а не StateEditingBio
}

func (m *Manager) Set(userID int64, state State) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.states[userID] = state
}

func (m *Manager) Reset(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.states, userID)
}
