package loaders

import (
	"context"

	"github.com/4udiwe/commnets-feed/internal/graph/model"
	"github.com/graph-gophers/dataloader/v7"
)

// CommentService интерфейс методов сервиса, необходимых для DataLoader
type CommentService interface {
	GetCommentsByPostBatch(ctx context.Context, postIDs []string) (map[string][]*model.Comment, error)
	GetChildrenBatch(ctx context.Context, parentIDs []string) (map[string][]*model.Comment, error)
}

// newCommentsByPostBatch создает batch function для DataLoader
func newCommentsByPostBatch(service CommentService) dataloader.BatchFunc[string, []*model.Comment] {
	return func(ctx context.Context, postIDs []string) []*dataloader.Result[[]*model.Comment] {
		results := make([]*dataloader.Result[[]*model.Comment], len(postIDs))

		// Получаем все комментарии для всех постов в одном запросе 
		batched, err := service.GetCommentsByPostBatch(ctx, postIDs)
		if err != nil {
			for i := range postIDs {
				results[i] = &dataloader.Result[[]*model.Comment]{
					Error: err,
				}
			}
			return results
		}

		// Распределяем результаты по порядку ID как в запросе
		for i, postID := range postIDs {
			results[i] = &dataloader.Result[[]*model.Comment]{
				Data: batched[postID],
			}
		}

		return results
	}
}

// newChildrenBatch создает batch function для DataLoader дочерних комментариев
func newChildrenBatch(service CommentService) dataloader.BatchFunc[string, []*model.Comment] {
	return func(ctx context.Context, parentIDs []string) []*dataloader.Result[[]*model.Comment] {
		results := make([]*dataloader.Result[[]*model.Comment], len(parentIDs))

		batched, err := service.GetChildrenBatch(ctx, parentIDs)
		if err != nil {
			for i := range parentIDs {
				results[i] = &dataloader.Result[[]*model.Comment]{
					Error: err,
				}
			}
			return results
		}

		for i, parentID := range parentIDs {
			results[i] = &dataloader.Result[[]*model.Comment]{
				Data: batched[parentID],
			}
		}

		return results
	}
}
