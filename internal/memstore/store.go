package memstore

import (
	"fmt"
	"log"
	newsv1 "news/buf/grpc/api/news/v1"
	"sync"
)

type MemStore interface {
	Create(news *newsv1.NewsServiceCreateResponse)
	Get(id string) (*newsv1.NewsServiceCreateResponse, error)
	GetAll() []*newsv1.NewsServiceCreateResponse
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

func (s *Store) GetAll() []*newsv1.NewsServiceCreateResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*newsv1.NewsServiceCreateResponse, 0, len(s.news))

	for _, news := range s.news {
		if news.DeletedAt == nil || news.DeletedAt.AsTime().IsZero(){
			result = append(result, news)
		}
	}

	return result
}

func (s *Store) Create(news *newsv1.NewsServiceCreateResponse) {
	s.mu.Lock()
	s.news[news.GetId()] = news
	log.Println("create news:", news.Id)
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
