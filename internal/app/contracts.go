package app

import (
	"context"

	"github.com/4udiwe/comments-feed/internal/graph/model"
)

// Общие контракты для репозиториев
// В зависимости от выбранном при запуске типе хранения данных их реализуют разные структуры
 
type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	GetByID(ctx context.Context, id string) (*model.Post, error)
	List(ctx context.Context, limit, offset int) ([]*model.Post, error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	GetByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error)
	GetChildren(ctx context.Context, parentID string, limit, offset int) ([]*model.Comment, error)
	GetByID(ctx context.Context, id string) (*model.Comment, error)
	GetCommentsByPostBatch(ctx context.Context, postIDs []string) (map[string][]*model.Comment, error)
	GetChildrenBatch(ctx context.Context, parentIDs []string) (map[string][]*model.Comment, error)
}
