package postgres_comment_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/4udiwe/comments-feed/internal/graph/model"
	"github.com/4udiwe/comments-feed/internal/repository"
	postgres_repository "github.com/4udiwe/comments-feed/internal/repository/postgres"
	"github.com/4udiwe/comments-feed/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type CommentRepository struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) *CommentRepository {
	return &CommentRepository{
		Postgres: pg,
	}
}

func (r *CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	query, args, _ := r.Builder.
		Insert("comment").
		Columns(
			"id",
			"post_id",
			"parent_id",
			"text",
			"created_at",
		).
		Values(
			comment.ID,
			comment.PostID,
			comment.ParentID,
			comment.Text,
			comment.CreatedAt,
		).
		Suffix("RETURNING id").
		ToSql()

	var id string

	err := r.GetTxManager(ctx).
		QueryRow(ctx, query, args...).
		Scan(&id)

	if err != nil {
		mapped := repository.MapPgError(err)

		logrus.WithFields(logrus.Fields{
			"id":      comment.ID,
			"post_id": comment.PostID,
		}).Warnf("failed to create comment: %v", err)

		return mapped
	}

	logrus.WithField("comment_id", id).Debug("comment created")
	return nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id string) (*model.Comment, error) {
	query, args, _ := r.Builder.
		Select(
			"id",
			"post_id",
			"parent_id",
			"text",
			"created_at",
		).
		From("comment").
		Where("id = ?", id).
		ToSql()

	row := r.GetTxManager(ctx).QueryRow(ctx, query, args...)

	commentDTO := &postgres_repository.CommentDTO{}
	err := row.Scan(
		&commentDTO.ID,
		&commentDTO.PostID,
		&commentDTO.ParentID,
		&commentDTO.Text,
		&commentDTO.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.WithField("comment_id", id).Debug("comment not found")
			return nil, repository.ErrCommentNotFound
		}

		logrus.WithError(err).WithField("comment_id", id).Error("failed to get comment")
		return nil, fmt.Errorf("failed to get comment: %w", repository.MapPgError(err))
	}

	logrus.WithField("comment_id", id).Debug("comment fetched")
	return commentDTO.ToEntity(), nil
}

func (r *CommentRepository) GetByPostID(
	ctx context.Context,
	postID string,
	limit, offset int,
) ([]*model.Comment, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	query, args, _ := r.Builder.
		Select(
			"id",
			"post_id",
			"parent_id",
			"text",
			"created_at",
		).
		From("comment").
		Where("post_id = ? AND parent_id IS NULL", postID).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	rows, err := r.GetTxManager(ctx).Query(ctx, query, args...)
	if err != nil {
		logrus.WithError(err).WithField("post_id", postID).Error("failed to get comments by post_id")
		return nil, fmt.Errorf("failed to get comments: %w", repository.MapPgError(err))
	}
	defer rows.Close()

	commentDTOs, err := pgx.CollectRows(rows, pgx.RowToStructByName[postgres_repository.CommentDTO])
	if err != nil {
		logrus.WithError(err).Error("failed to scan comments")
		return nil, fmt.Errorf("failed to scan comments: %w", err)
	}

	if len(commentDTOs) == 0 {
		return []*model.Comment{}, nil
	}

	comments := lo.Map(commentDTOs, func(dto postgres_repository.CommentDTO, _ int) *model.Comment {
		return dto.ToEntity()
	})

	logrus.WithField("count", len(comments)).Debug("comments fetched by post")
	return comments, nil
}

func (r *CommentRepository) GetChildren(
	ctx context.Context,
	parentID string,
	limit, offset int,
) ([]*model.Comment, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	query, args, _ := r.Builder.
		Select(
			"id",
			"post_id",
			"parent_id",
			"text",
			"created_at",
		).
		From("comment").
		Where("parent_id = ?", parentID).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	rows, err := r.GetTxManager(ctx).Query(ctx, query, args...)
	if err != nil {
		logrus.WithError(err).WithField("parent_id", parentID).Error("failed to get child comments")
		return nil, fmt.Errorf("failed to get child comments: %w", repository.MapPgError(err))
	}
	defer rows.Close()

	commentDTOs, err := pgx.CollectRows(rows, pgx.RowToStructByName[postgres_repository.CommentDTO])
	if err != nil {
		logrus.WithError(err).Error("failed to scan child comments")
		return nil, fmt.Errorf("failed to scan child comments: %w", err)
	}

	if len(commentDTOs) == 0 {
		return []*model.Comment{}, nil
	}

	comments := lo.Map(commentDTOs, func(dto postgres_repository.CommentDTO, _ int) *model.Comment {
		return dto.ToEntity()
	})

	logrus.WithField("count", len(comments)).Debug("children fetched")
	return comments, nil
}

// GetCommentsByPostBatch получает комментарии для множества постов в одном SQL запросе
// Используется DataLoader
// Возвращает map[postID] -> []*Comment
func (r *CommentRepository) GetCommentsByPostBatch(
	ctx context.Context,
	postIDs []string,
) (map[string][]*model.Comment, error) {
	if len(postIDs) == 0 {
		return make(map[string][]*model.Comment), nil
	}

	// Построить placeholders для IN clause
	var placeholders []interface{}
	var placeholderStr string
	for i, id := range postIDs {
		if i > 0 {
			placeholderStr += ", "
		}
		placeholderStr += fmt.Sprintf("$%d", i+1)
		placeholders = append(placeholders, id)
	}

	queryStr := fmt.Sprintf(`
		SELECT id, post_id, parent_id, text, created_at 
		FROM comment 
		WHERE post_id IN (%s) AND parent_id IS NULL 
		ORDER BY created_at DESC
	`, placeholderStr)

	rows, err := r.GetTxManager(ctx).Query(ctx, queryStr, placeholders...)
	if err != nil {
		logrus.WithError(err).WithField("post_ids", postIDs).Error("failed to batch load comments by post")
		return nil, fmt.Errorf("failed to batch load comments: %w", repository.MapPgError(err))
	}
	defer rows.Close()

	// Инициализировать результат для всех postID
	result := make(map[string][]*model.Comment)
	for _, postID := range postIDs {
		result[postID] = []*model.Comment{}
	}

	commentDTOs, err := pgx.CollectRows(rows, pgx.RowToStructByName[postgres_repository.CommentDTO])
	if err != nil {
		logrus.WithError(err).Error("failed to scan batch comments")
		return nil, fmt.Errorf("failed to scan batch comments: %w", err)
	}

	// Группировать комментарии по postID
	for _, dto := range commentDTOs {
		result[dto.PostID] = append(result[dto.PostID], dto.ToEntity())
	}

	logrus.WithField("count", len(commentDTOs)).WithField("post_ids_count", len(postIDs)).Debug("batch loaded comments by post")
	return result, nil
}

// GetChildrenBatch получает дочерние комментарии для множества родителей в одном запросе
// Используется DataLoader
// Возвращает map[parentID] -> []*Comment
func (r *CommentRepository) GetChildrenBatch(
	ctx context.Context,
	parentIDs []string,
) (map[string][]*model.Comment, error) {
	if len(parentIDs) == 0 {
		return make(map[string][]*model.Comment), nil
	}

	// Построить placeholders для IN clause
	var placeholders []interface{}
	var placeholderStr string
	for i, id := range parentIDs {
		if i > 0 {
			placeholderStr += ", "
		}
		placeholderStr += fmt.Sprintf("$%d", i+1)
		placeholders = append(placeholders, id)
	}

	queryStr := fmt.Sprintf(`
		SELECT id, post_id, parent_id, text, created_at 
		FROM comment 
		WHERE parent_id IN (%s) 
		ORDER BY created_at DESC
	`, placeholderStr)

	rows, err := r.GetTxManager(ctx).Query(ctx, queryStr, placeholders...)
	if err != nil {
		logrus.WithError(err).WithField("parent_ids", parentIDs).Error("failed to batch load children comments")
		return nil, fmt.Errorf("failed to batch load children: %w", repository.MapPgError(err))
	}
	defer rows.Close()

	// Инициализировать результат для всех parentID
	result := make(map[string][]*model.Comment)
	for _, parentID := range parentIDs {
		result[parentID] = []*model.Comment{}
	}

	commentDTOs, err := pgx.CollectRows(rows, pgx.RowToStructByName[postgres_repository.CommentDTO])
	if err != nil {
		logrus.WithError(err).Error("failed to scan batch children")
		return nil, fmt.Errorf("failed to scan batch children: %w", err)
	}

	// Группировать комментарии по parentID
	for _, dto := range commentDTOs {
		if dto.ParentID != nil {
			result[*dto.ParentID] = append(result[*dto.ParentID], dto.ToEntity())
		}
	}

	logrus.WithField("count", len(commentDTOs)).WithField("parent_ids_count", len(parentIDs)).Debug("batch loaded children comments")
	return result, nil
}
