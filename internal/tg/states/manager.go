package states

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type Manager struct {
	states map[int64]State
	redis  *redis.Client
	ctx    context.Context
	ttl    time.Duration
	mu     sync.Mutex
}

func NewManager(redisClient *redis.Client) *Manager {
	return &Manager{
		states: make(map[int64]State),
		redis:  redisClient,
		ctx:    context.Background(),
		ttl:    24 * time.Hour,
	}
}

func (m *Manager) Get(userID int64) State {
	m.mu.Lock()
	defer m.mu.Unlock()
	if state, exists := m.states[userID]; exists {
		return state
	}
	return StateDefault
}

func (m *Manager) Set(userID int64, state State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[userID] = state
}

func (m *Manager) Reset(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[userID] = StateDefault
}

func (m *Manager) GetCurrentIndex(userID int64) int {
	indexStr, err := m.redis.Get(m.ctx, m.indexKey(userID, "search")).Result()
	if err == redis.Nil {
		return 0
	} else if err != nil {
		return 0
	}
	index, _ := strconv.Atoi(indexStr)
	return index
}

func (m *Manager) SetCurrentIndex(userID int64, index int) {
	m.redis.Set(m.ctx, m.indexKey(userID, "search"), index, m.ttl)
}

func (m *Manager) ResetCurrentIndex(userID int64) {
	m.redis.Set(m.ctx, m.indexKey(userID, "search"), 0, m.ttl)
}

func (m *Manager) GetLikesCurrentIndex(userID int64) int {
	indexStr, err := m.redis.Get(m.ctx, m.indexKey(userID, "likes")).Result()
	if err == redis.Nil {
		return 0
	} else if err != nil {
		return 0
	}
	index, _ := strconv.Atoi(indexStr)
	return index
}

func (m *Manager) SetLikesCurrentIndex(userID int64, index int) {
	m.redis.Set(m.ctx, m.indexKey(userID, "likes"), index, m.ttl)
}

func (m *Manager) ResetLikesCurrentIndex(userID int64) {
	m.redis.Set(m.ctx, m.indexKey(userID, "likes"), 0, m.ttl)
}

func (m *Manager) indexKey(userID int64, prefix string) string {
	return prefix + "_index:user:" + strconv.FormatInt(userID, 10)
}
