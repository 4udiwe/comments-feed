package post_service

import (
	"context"

	"github.com/4udiwe/commnets-feed/internal/graph/model"
)

//go:generate go tool mockgen -source=contracts.go -destination=mocks/mocks.go

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	GetByID(ctx context.Context, id string) (*model.Post, error)
	List(ctx context.Context, limit, offset int) ([]*model.Post, error)
}
