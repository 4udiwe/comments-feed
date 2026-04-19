package memory_post_repository

import (
	"context"
	"sync"

	"github.com/4udiwe/comments-feed/internal/graph/model"
	"github.com/4udiwe/comments-feed/internal/repository"
	"github.com/sirupsen/logrus"
)

type PostRepository struct {
	mu    sync.RWMutex
	posts []*model.Post
	byID  map[string]*model.Post
}

func NewPostRepository() *PostRepository {
	return &PostRepository{
		posts: []*model.Post{},
		byID:  make(map[string]*model.Post),
	}
}

func (r *PostRepository) Create(ctx context.Context, post *model.Post) error {
		r.mu.Lock()
	defer r.mu.Unlock()

	// Check if post already exists
	if _, exists := r.byID[post.ID]; exists {
		logrus.WithField("post_id", post.ID).Warn("post already exists")
		return repository.ErrDuplicateKey
	}

	r.posts = append(r.posts, post)
	r.byID[post.ID] = post

	logrus.WithField("post_id", post.ID).Debug("post created in memory")
	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, ok := r.byID[id]
	if !ok {
		logrus.WithField("post_id", id).Debug("post not found")
		return nil, repository.ErrPostNotFound
	}

	return post, nil
}

func (r *PostRepository) List(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.posts) == 0 {
		logrus.Debug("no posts found")
		return []*model.Post{}, nil
	}

	if offset >= len(r.posts) {
		return []*model.Post{}, nil
	}

	end := offset + limit
	if end > len(r.posts) {
		end = len(r.posts)
	}

	return r.posts[offset:end], nil
}
