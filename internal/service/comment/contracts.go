package comment_service

import (
	"context"

	"github.com/4udiwe/comments-feed/internal/broker"
	"github.com/4udiwe/comments-feed/internal/graph/model"
)

//go:generate go tool mockgen -source=contracts.go -destination=mocks/mocks.go

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	GetByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error)
	GetChildren(ctx context.Context, parentID string, limit, offset int) ([]*model.Comment, error)
	GetByID(ctx context.Context, id string) (*model.Comment, error)
	
	// Batch методы для DataLoader
	GetCommentsByPostBatch(ctx context.Context, postIDs []string) (map[string][]*model.Comment, error)
	GetChildrenBatch(ctx context.Context, parentIDs []string) (map[string][]*model.Comment, error)
}

type PostRepository interface {
	GetByID(ctx context.Context, id string) (*model.Post, error)
}

type CommentBroker interface {
	Publish(postID string, comment *broker.Comment)
	Subscribe(postID string) (<-chan *broker.Comment, func())
}
