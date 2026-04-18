package comment_service

import (
	"context"
	"strings"
	"time"

	"github.com/4udiwe/commnets-feed/internal/broker"
	"github.com/4udiwe/commnets-feed/internal/graph/errors"
	"github.com/4udiwe/commnets-feed/internal/graph/model"
	"github.com/4udiwe/commnets-feed/pkg/transactor"
	"github.com/google/uuid"
)

type CommentService struct {
	commentRepo   CommentRepository
	commentBroker CommentBroker
	postRepo      PostRepository
	txManager     transactor.Transactor
}

func NewCommentService(
	commentRepo CommentRepository,
	commentBroker CommentBroker,
	postRepo PostRepository,
	txManager transactor.Transactor,
) *CommentService {
	return &CommentService{
		commentRepo:   commentRepo,
		commentBroker: commentBroker,
		postRepo:      postRepo,
		txManager:     txManager,
	}
}

func (s *CommentService) CreateComment(ctx context.Context, input model.CreateCommentInput) (*model.Comment, error) {
	// 1. валидация input'а
	validator := errors.NewInputValidator()

	// Валидируем ID поста
	if err := validator.ValidateID(input.PostID, "postId"); err != nil {
		return nil, err
	}

	// Валидируем текст коммента
	trimmedText := strings.TrimSpace(input.Text)
	if err := validator.ValidateString(trimmedText, "text", 1, 2000); err != nil {
		return nil, err
	}

	// Если parentId установлен, валидируем его
	if input.ParentID.IsSet() && input.ParentID.Value() != nil {
		if err := validator.ValidateID(*input.ParentID.Value(), "parentId"); err != nil {
			return nil, err
		}
	}

	var comment *model.Comment

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// 2. проверка поста
		post, err := s.postRepo.GetByID(ctx, input.PostID)
		if err != nil {
			return errors.NewNotFoundError("post")
		}

		// 3. проверка commentsEnabled
		if !post.CommentsEnabled {
			return errors.NewForbiddenError("comments are disabled for this post")
		}

		var parentID *string

		// 4. если есть parent
		if input.ParentID.IsSet() && input.ParentID.Value() != nil {
			parentID = input.ParentID.Value()

			parent, err := s.commentRepo.GetByID(ctx, *parentID)
			if err != nil {
				return errors.NewNotFoundError("parent comment")
			}

			// защита от кривой вложенности
			if parent.PostID != input.PostID {
				return errors.NewBadRequestError("parent comment belongs to another post")
			}
		}

		comment = &model.Comment{
			ID:        uuid.NewString(),
			PostID:    input.PostID,
			ParentID:  parentID,
			Text:      trimmedText,
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		// 5. создание коммента
		if err := s.commentRepo.Create(ctx, comment); err != nil {
			return errors.NewInternalError("cannot create comment", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	/*
		На данный момент публикация событий происходит не атомарно.
		При добавлении внешнего брокера сообщений (Kafka/Redis) следует использовать
		Outbox-pattern и перенести создание события в транзакцию.
	*/

	// 6. публикация события
	s.commentBroker.Publish(comment.PostID,
		&broker.Comment{
			ID:      comment.ID,
			Content: comment.Text,
			PostID:  comment.PostID,
		},
	)

	return comment, nil
}

func (s *CommentService) GetCommentsByPost(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {
	validator := errors.NewInputValidator()

	// Валидируем входные параметры
	if err := validator.ValidateID(postID, "postId"); err != nil {
		return nil, err
	}

	if err := validator.ValidateInt(int32(limit), "limit", 1, 100); err != nil {
		return nil, err
	}

	if err := validator.ValidateInt(int32(offset), "offset", 0, 1000000); err != nil {
		return nil, err
	}

	comments, err := s.commentRepo.GetByPostID(ctx, postID, limit, offset)

	if err != nil {
		return nil, errors.NewInternalError("cannot fetch comments", err)
	}

	return comments, nil
}

func (s *CommentService) GetChildren(ctx context.Context, parentID string, limit, offset int) ([]*model.Comment, error) {
	validator := errors.NewInputValidator()

	// Валидируем входные параметры
	if err := validator.ValidateID(parentID, "parentId"); err != nil {
		return nil, err
	}

	if err := validator.ValidateInt(int32(limit), "limit", 1, 100); err != nil {
		return nil, err
	}

	if err := validator.ValidateInt(int32(offset), "offset", 0, 1000000); err != nil {
		return nil, err
	}

	comments, err := s.commentRepo.GetChildren(ctx, parentID, limit, offset)

	if err != nil {
		return nil, errors.NewInternalError("cannot fetch comments", err)
	}

	return comments, nil
}

func (s *CommentService) SubscribeToComments(ctx context.Context, postID string) (<-chan *model.Comment, error) {
	ch, unsubscribe := s.commentBroker.Subscribe(postID)

	out := make(chan *model.Comment, 1)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				unsubscribe()
				return

			case msg := <-ch:
				out <- &model.Comment{
					ID:     msg.ID,
					Text:   msg.Content,
					PostID: msg.PostID,
				}
			}
		}
	}()

	return out, nil
}

// GetCommentsByPostWithPagination получает комментарии к посту с пагинацией
// Предназначен для использования из резолверов с DataLoader
func (s *CommentService) GetCommentsByPostWithPagination(
	ctx context.Context,
	postID string,
	limit, offset int,
) ([]*model.Comment, error) {
	validator := errors.NewInputValidator()

	// Валидируем параметры
	if err := validator.ValidateID(postID, "postId"); err != nil {
		return nil, err
	}

	if err := validator.ValidateInt(int32(limit), "limit", 1, 100); err != nil {
		return nil, err
	}

	if err := validator.ValidateInt(int32(offset), "offset", 0, 1000000); err != nil {
		return nil, err
	}

	// Получить ВСЕ комментарии для этого поста
	allComments, err := s.commentRepo.GetCommentsByPostBatch(ctx, []string{postID})
	if err != nil {
		return nil, errors.NewInternalError("cannot fetch comments", err)
	}

	comments := allComments[postID]
	if len(comments) == 0 {
		return []*model.Comment{}, nil
	}

	// Применить пагинацию локально
	if offset >= len(comments) {
		return []*model.Comment{}, nil
	}

	end := offset + limit
	if end > len(comments) {
		end = len(comments)
	}

	return comments[offset:end], nil
}

// GetChildrenWithPagination получает дочерние комментарии с пагинацией
// Предназначен для использования из резолверов с DataLoader
func (s *CommentService) GetChildrenWithPagination(
	ctx context.Context,
	parentID string,
	limit, offset int,
) ([]*model.Comment, error) {
	validator := errors.NewInputValidator()

	// Валидируем параметры
	if err := validator.ValidateID(parentID, "parentId"); err != nil {
		return nil, err
	}

	if err := validator.ValidateInt(int32(limit), "limit", 1, 100); err != nil {
		return nil, err
	}

	if err := validator.ValidateInt(int32(offset), "offset", 0, 1000000); err != nil {
		return nil, err
	}

	// Получить ВСЕ дочерние комментарии
	allChildren, err := s.commentRepo.GetChildrenBatch(ctx, []string{parentID})
	if err != nil {
		return nil, errors.NewInternalError("cannot fetch comments", err)
	}

	children := allChildren[parentID]
	if len(children) == 0 {
		return []*model.Comment{}, nil
	}

	// Применить пагинацию локально
	if offset >= len(children) {
		return []*model.Comment{}, nil
	}

	end := offset + limit
	if end > len(children) {
		end = len(children)
	}

	return children[offset:end], nil
}

// GetCommentsByPostBatch получает все комментарии для нескольких постов в одном запросе
// Используется DataLoader
func (s *CommentService) GetCommentsByPostBatch(
	ctx context.Context,
	postIDs []string,
) (map[string][]*model.Comment, error) {
	return s.commentRepo.GetCommentsByPostBatch(ctx, postIDs)
}

// GetChildrenBatch получает дочерние комментарии для нескольких родительских комментариев в одном запросе
// Используется DataLoader
func (s *CommentService) GetChildrenBatch(
	ctx context.Context,
	parentIDs []string,
) (map[string][]*model.Comment, error) {
	return s.commentRepo.GetChildrenBatch(ctx, parentIDs)
}
