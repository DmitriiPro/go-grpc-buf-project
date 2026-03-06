package memstore

import (
	"fmt"
	newsv1 "news/buf/grpc/api/news/v1"
	"sync"
)

type MemStore interface {
	Create(news *newsv1.NewsServiceCreateResponse)
	Get(id string) (*newsv1.NewsServiceCreateResponse, error)
}

type Store struct {
	mu   sync.RWMutex
	news map[string]*newsv1.NewsServiceCreateResponse
}

func NewStore() *Store {
	return &Store{
		news: make(map[string]*newsv1.NewsServiceCreateResponse),
	}
}

func (s *Store) Create(news *newsv1.NewsServiceCreateResponse) {
	s.mu.Lock()
	s.news[news.GetId()] = news
	s.mu.Unlock()
}

func (s *Store) Get(id string) (*newsv1.NewsServiceCreateResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if news, exists := s.news[id]; exists {
		return news, nil
	}
	return nil, fmt.Errorf("news %s not found", id)
}