package memstore

import (
	"fmt"
	"log"
	newsv1 "news/buf/grpc/api/news/v1"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MemStore interface {
	Create(news *newsv1.NewsServiceCreateResponse)
	Get(id string) (*newsv1.NewsServiceCreateResponse, error)
	GetAll() []*newsv1.NewsServiceCreateResponse
	UpdateNews(*newsv1.NewsServiceCreateRequest)
	DeleteNews(id uuid.UUID)
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

func (s *Store) DeleteNews(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for idx, news := range s.news {
		if news.Id == id.String() {
			s.news[idx].DeletedAt = timestamppb.Now()
			return
		}
	}
}

func (s *Store) UpdateNews(req *newsv1.NewsServiceCreateRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for idx, news := range s.news {
		if news.Id == req.Id && news.DeletedAt == nil {
			s.news[idx] = &newsv1.NewsServiceCreateResponse{
				Id:        req.GetId(),
				Author:    req.GetAuthor(),
				Title:     req.GetTitle(),
				Summary:   req.GetSummary(),
				Content:   req.GetContent(),
				Source:    req.GetSource(),
				Tags:      req.GetTags(),
				CreatedAt: news.CreatedAt,
				UpdatedAt: timestamppb.Now(),
				DeletedAt: news.DeletedAt,
			}
			return
		}
	}

	if news, exists := s.news[req.GetId()]; exists {
		news.Author = req.GetAuthor()
		news.Title = req.GetTitle()
		news.Summary = req.GetSummary()
		news.Content = req.GetContent()
		news.Source = req.GetSource()
		news.Tags = req.GetTags()
		news.UpdatedAt = timestamppb.Now()
	}

}

func (s *Store) GetAll() []*newsv1.NewsServiceCreateResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*newsv1.NewsServiceCreateResponse, 0, len(s.news))

	for _, news := range s.news {
		if news.DeletedAt == nil || news.DeletedAt.AsTime().IsZero() {
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
