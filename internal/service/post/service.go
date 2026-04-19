package post_service

import (
	"context"
	"errors"
	"strings"
	"time"

	gqlerrors "github.com/4udiwe/comments-feed/internal/graph/errors"
	"github.com/4udiwe/comments-feed/internal/graph/model"
	"github.com/4udiwe/comments-feed/internal/repository"
	"github.com/google/uuid"
)

type PostService struct {
	repo      PostRepository
	validator gqlerrors.InputValidator
}

func NewPostService(repo PostRepository, validator gqlerrors.InputValidator) *PostService {
	return &PostService{
		repo:      repo,
		validator: validator,
	}
}

func (s *PostService) CreatePost(ctx context.Context, input model.CreatePostInput) (*model.Post, error) {
	// Валидируем и обрезаем пробелы из title
	trimmedTitle := strings.TrimSpace(input.Title)
	if err := s.validator.ValidateString(trimmedTitle, "title", 1, 500); err != nil {
		return nil, err
	}

	// Валидируем и обрезаем пробелы из content
	trimmedContent := strings.TrimSpace(input.Content)
	if err := s.validator.ValidateString(trimmedContent, "content", 1, 10000); err != nil {
		return nil, err
	}

	commentsEnabled := true
	if input.CommentsEnabled.IsSet() && input.CommentsEnabled.Value() != nil {
		commentsEnabled = *input.CommentsEnabled.Value()
	}

	post := &model.Post{
		ID:              uuid.New().String(),
		Title:           trimmedTitle,
		Content:         trimmedContent,
		CommentsEnabled: commentsEnabled,
		CreatedAt:       time.Now().Format(time.RFC3339),
	}

	if err := s.repo.Create(ctx, post); err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			return nil, gqlerrors.NewConflictError("post with this title already exists")
		}
		return nil, gqlerrors.NewInternalError("cannot create post", err)
	}

	return post, nil
}

func (s *PostService) GetPost(ctx context.Context, id string) (*model.Post, error) {
	// Валидируем ID
	if err := s.validator.ValidateID(id, "id"); err != nil {
		return nil, err
	}

	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			return nil, gqlerrors.NewNotFoundError("post")
		}
		return nil, gqlerrors.NewInternalError("cannot fetch post", err)
	}
	return post, nil
}

func (s *PostService) ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	// Валидируем параметры пагинации
	if err := s.validator.ValidateInt(int32(limit), "limit", 1, 100); err != nil {
		return nil, err
	}

	if err := s.validator.ValidateInt(int32(offset), "offset", 0, 1000000); err != nil {
		return nil, err
	}

	posts, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, gqlerrors.NewInternalError("cannot fetch posts", err)
	}
	return posts, nil
}
