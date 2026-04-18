package post_service

import (
	"context"
	stderrors "errors"
	"strings"
	"time"

	"github.com/4udiwe/commnets-feed/internal/graph/errors"
	"github.com/4udiwe/commnets-feed/internal/graph/model"
	"github.com/4udiwe/commnets-feed/internal/repository"
	"github.com/google/uuid"
)

type PostService struct {
	repo PostRepository
}

func NewPostService(repo PostRepository) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) CreatePost(ctx context.Context, input model.CreatePostInput) (*model.Post, error) {
	// Валидируем input
	validator := errors.NewInputValidator()

	// Валидируем и обрезаем пробелы из title
	trimmedTitle := strings.TrimSpace(input.Title)
	if err := validator.ValidateString(trimmedTitle, "title", 1, 500); err != nil {
		return nil, err
	}

	// Валидируем и обрезаем пробелы из content
	trimmedContent := strings.TrimSpace(input.Content)
	if err := validator.ValidateString(trimmedContent, "content", 1, 10000); err != nil {
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
		if stderrors.Is(err, repository.ErrDuplicateKey) {
			return nil, errors.NewConflictError("post with this title already exists")
		}
		return nil, errors.NewInternalError("cannot create post", err)
	}

	return post, nil
}

func (s *PostService) GetPost(ctx context.Context, id string) (*model.Post, error) {
	// Валидируем ID
	validator := errors.NewInputValidator()
	if err := validator.ValidateID(id, "id"); err != nil {
		return nil, err
	}

	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if stderrors.Is(err, repository.ErrPostNotFound) {
			return nil, errors.NewNotFoundError("post")
		}
		return nil, errors.NewInternalError("cannot fetch post", err)
	}
	return post, nil
}

func (s *PostService) ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	// Валидируем параметры пагинации
	validator := errors.NewInputValidator()
	if err := validator.ValidateInt(int32(limit), "limit", 1, 100); err != nil {
		return nil, err
	}

	if err := validator.ValidateInt(int32(offset), "offset", 0, 1000000); err != nil {
		return nil, err
	}

	posts, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, errors.NewInternalError("cannot fetch posts", err)
	}
	return posts, nil
}
