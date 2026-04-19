package postgres_post_repository

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

type PostRepository struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) *PostRepository {
	return &PostRepository{
		Postgres: pg,
	}
}

func (r *PostRepository) Create(ctx context.Context, post *model.Post) error {
	query, args, _ := r.Builder.
		Insert("post").
		Columns(
			"id",
			"title",
			"content",
			"comments_enabled",
			"created_at",
		).
		Values(
			post.ID,
			post.Title,
			post.Content,
			post.CommentsEnabled,
			post.CreatedAt,
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
			"id":    post.ID,
			"title": post.Title,
		}).Warnf("failed to create post: %v", err)

		return mapped
	}

	logrus.WithField("post_id", id).Debug("post created")
	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	query, args, _ := r.Builder.
		Select(
			"id",
			"title",
			"content",
			"comments_enabled",
			"created_at",
		).
		From("post").
		Where("id = ?", id).
		ToSql()

	row := r.GetTxManager(ctx).QueryRow(ctx, query, args...)

	postDTO := &postgres_repository.PostDTO{}
	err := row.Scan(
		&postDTO.ID,
		&postDTO.Title,
		&postDTO.Content,
		&postDTO.CommentsEnabled,
		&postDTO.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.WithField("post_id", id).Debug("post not found")
			return nil, repository.ErrPostNotFound
		}

		logrus.WithError(err).WithField("post_id", id).Error("failed to get post")
		return nil, fmt.Errorf("failed to get post: %w", repository.MapPgError(err))
	}

	logrus.WithField("post_id", id).Debug("post fetched")
	return postDTO.ToEntity(), nil
}

func (r *PostRepository) List(ctx context.Context, limit int, offset int) ([]*model.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	query, args, _ := r.Builder.
		Select(
			"id",
			"title",
			"content",
			"comments_enabled",
			"created_at",
		).
		From("post").
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	rows, err := r.GetTxManager(ctx).Query(ctx, query, args...)
	if err != nil {
		logrus.WithError(err).Error("failed to list posts")
		return nil, fmt.Errorf("failed to list posts: %w", repository.MapPgError(err))
	}
	defer rows.Close()

	postDTOs, err := pgx.CollectRows(rows, pgx.RowToStructByName[postgres_repository.PostDTO])
	if err != nil {
		logrus.WithError(err).Error("failed to scan posts")
		return nil, fmt.Errorf("failed to scan posts: %w", err)
	}

	if len(postDTOs) == 0 {
		return []*model.Post{}, nil
	}

	posts := lo.Map(postDTOs, func(dto postgres_repository.PostDTO, _ int) *model.Post {
		return dto.ToEntity()
	})

	logrus.WithField("count", len(posts)).Debug("posts fetched")
	return posts, nil
}

