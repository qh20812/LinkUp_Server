package services

import (
	"context"
	"fmt"
	"strings"

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
		return resp, nil
	}

	if err != nil {
		return dto.SearchResponse{}, fmt.Errorf("search %s: %w", searchType, err)
	}

	return resp, nil
}
