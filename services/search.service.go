package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"linkup/dto"
	"linkup/repository"
	"linkup/validations"
)

type SearchService struct {
	searchRepo *repository.SearchRepository
	validation *validations.SearchValidation
}

func NewSearchService(searchRepo *repository.SearchRepository, validation *validations.SearchValidation) *SearchService {
	return &SearchService{
		searchRepo: searchRepo,
		validation: validation,
	}
}

type trendingCache struct {
	mu     sync.RWMutex
	data   []dto.HashtagSearchResult
	expiry time.Time
}

var trendingCacheInstance = &trendingCache{}

const trendingCacheTTL = 5 * time.Minute

func (s *SearchService) GetTrendingHashtags(ctx context.Context) ([]dto.HashtagSearchResult, error) {
	trendingCacheInstance.mu.RLock()
	valid := time.Now().Before(trendingCacheInstance.expiry)
	hasCached := len(trendingCacheInstance.data) > 0
	if valid {
		data := trendingCacheInstance.data
		trendingCacheInstance.mu.RUnlock()
		return data, nil
	}
	trendingCacheInstance.mu.RUnlock()

	trendingCacheInstance.mu.Lock()
	defer trendingCacheInstance.mu.Unlock()

	if time.Now().Before(trendingCacheInstance.expiry) {
		return trendingCacheInstance.data, nil
	}

	data, err := s.searchRepo.GetTrendingHashtags(ctx)
	if err != nil {
		if hasCached {
			return trendingCacheInstance.data, nil
		}
		return nil, err
	}

	trendingCacheInstance.data = data
	trendingCacheInstance.expiry = time.Now().Add(trendingCacheTTL)
	return data, nil
}

func (s *SearchService) Search(ctx context.Context, input dto.SearchInput) (dto.SearchResponse, error) {
	searchType := strings.TrimSpace(input.Type)
	if searchType == "" {
		searchType = "all"
	}

	if err := s.validation.ValidateSearch(input.Keyword, searchType); err != nil {
		return dto.SearchResponse{}, err
	}

	var resp dto.SearchResponse
	var err error

	switch searchType {
	case "users":
		resp.Users, err = s.searchRepo.SearchUsers(ctx, input.Keyword)
	case "posts":
		resp.Posts, err = s.searchRepo.SearchPosts(ctx, input.Keyword)
	case "hashtags":
		resp.Hashtags, err = s.searchRepo.SearchHashtags(ctx, input.Keyword)
	default:
		resp.Users, err = s.searchRepo.SearchUsers(ctx, input.Keyword)
		if err != nil {
			return dto.SearchResponse{}, fmt.Errorf("search users: %w", err)
		}
		resp.Posts, err = s.searchRepo.SearchPosts(ctx, input.Keyword)
		if err != nil {
			return dto.SearchResponse{}, fmt.Errorf("search posts: %w", err)
		}
		resp.Hashtags, err = s.searchRepo.SearchHashtags(ctx, input.Keyword)
		if err != nil {
			return dto.SearchResponse{}, fmt.Errorf("search hashtags: %w", err)
		}
		setSearchMessage(&resp)
		return resp, nil
	}

	if err != nil {
		return dto.SearchResponse{}, fmt.Errorf("search %s: %w", searchType, err)
	}

	setSearchMessage(&resp)
	return resp, nil
}

func setSearchMessage(resp *dto.SearchResponse) {
	if len(resp.Users) == 0 && len(resp.Posts) == 0 && len(resp.Hashtags) == 0 {
		resp.Message = "Không tìm thấy kết quả phù hợp"
	}
}
