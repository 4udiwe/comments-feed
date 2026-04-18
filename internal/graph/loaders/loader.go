package loaders

import (
	"context"

	"github.com/4udiwe/commnets-feed/internal/graph/model"
	"github.com/graph-gophers/dataloader/v7"
)

type contextKey string

const (
	commentsByPostLoaderKey contextKey = "commentsByPostLoader"
	childrenLoaderKey       contextKey = "childrenLoader"
)

// Loaders объект содержит все DataLoaders для текущего GraphQL запроса
// Каждый GraphQL запрос должен иметь свой собственный экземпляр Loaders
type Loaders struct {
	// CommentsByPostLoader батчирует загрузку комментариев по postID
	CommentsByPostLoader *dataloader.Loader[string, []*model.Comment]

	// ChildrenLoader батчирует загрузку дочерних комментариев по parentID
	ChildrenLoader *dataloader.Loader[string, []*model.Comment]
}

// AttachLoaders добавляет Loaders в context
// Вызывается из middleware при начале каждого GraphQL запроса
func AttachLoaders(ctx context.Context, loaders *Loaders) context.Context {
	ctx = context.WithValue(ctx, commentsByPostLoaderKey, loaders.CommentsByPostLoader)
	ctx = context.WithValue(ctx, childrenLoaderKey, loaders.ChildrenLoader)
	return ctx
}

// GetCommentsByPostLoader извлекает DataLoader для комментариев по постам из context
func GetCommentsByPostLoader(ctx context.Context) *dataloader.Loader[string, []*model.Comment] {
	loader, ok := ctx.Value(commentsByPostLoaderKey).(*dataloader.Loader[string, []*model.Comment])
	if !ok {
		return nil
	}
	return loader
}

// GetChildrenLoader извлекает DataLoader для дочерних комментариев из context
func GetChildrenLoader(ctx context.Context) *dataloader.Loader[string, []*model.Comment] {
	loader, ok := ctx.Value(childrenLoaderKey).(*dataloader.Loader[string, []*model.Comment])
	if !ok {
		return nil
	}
	return loader
}

// NewLoaders создает новый набор DataLoaders для текущего GraphQL запроса
// Вызывается из middleware
func NewLoaders(commentService CommentService) *Loaders {
	return &Loaders{
		CommentsByPostLoader: dataloader.NewBatchedLoader[string, []*model.Comment](
			newCommentsByPostBatch(commentService),
		),
		ChildrenLoader: dataloader.NewBatchedLoader[string, []*model.Comment](
			newChildrenBatch(commentService),
		),
	}
}
