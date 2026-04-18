package memory_comment_repository

import (
	"context"
	"sync"

	"github.com/4udiwe/commnets-feed/internal/graph/model"
	"github.com/4udiwe/commnets-feed/internal/repository"
	"github.com/sirupsen/logrus"
)

type CommentRepository struct {
	mu sync.RWMutex

	byID       map[string]*model.Comment
	byPostID   map[string][]*model.Comment
	byParentID map[string][]*model.Comment
}

func NewCommentRepository() *CommentRepository {
	return &CommentRepository{
		byID:       make(map[string]*model.Comment),
		byPostID:   make(map[string][]*model.Comment),
		byParentID: make(map[string][]*model.Comment),
	}
}

func (r *CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// проверка на unique id
	if _, exists := r.byID[comment.ID]; exists {
		logrus.WithField("comment_id", comment.ID).Warn("comment already exists")
		return repository.ErrDuplicateKey
	}

	r.byID[comment.ID] = comment

	// индекс по postID
	r.byPostID[comment.PostID] = append(r.byPostID[comment.PostID], comment)

	// индекс по parentID (только если есть)
	if comment.ParentID != nil {
		parentID := *comment.ParentID
		r.byParentID[parentID] = append(r.byParentID[parentID], comment)
	}

	logrus.WithField("comment_id", comment.ID).Debug("comment created in memory")
	return nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id string) (*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	comment, ok := r.byID[id]
	if !ok {
		logrus.WithField("comment_id", id).Debug("comment not found")
		return nil, repository.ErrCommentNotFound
	}

	return comment, nil
}

func (r *CommentRepository) GetByPostID(
	ctx context.Context,
	postID string,
	limit,
	offset int,
) ([]*model.Comment, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	all := r.byPostID[postID]

	// фильтруем только root (parentID == nil)
	var roots []*model.Comment
	for _, c := range all {
		if c.ParentID == nil {
			roots = append(roots, c)
		}
	}

	if len(roots) == 0 {
		logrus.WithField("post_id", postID).Debug("no comments found for post")
		return []*model.Comment{}, nil
	}

	if offset >= len(roots) {
		return []*model.Comment{}, nil
	}

	end := offset + limit
	if end > len(roots) {
		end = len(roots)
	}

	return roots[offset:end], nil
}

func (r *CommentRepository) GetChildren(
	ctx context.Context,
	parentID string,
	limit,
	offset int,
) ([]*model.Comment, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	children := r.byParentID[parentID]

	if len(children) == 0 {
		return []*model.Comment{}, nil
	}

	if offset >= len(children) {
		return []*model.Comment{}, nil
	}

	end := offset + limit
	if end > len(children) {
		end = len(children)
	}

	return children[offset:end], nil
}

// GetCommentsByPostBatch получает комментарии для множества постов
// Возвращает map[postID] -> []*Comment (только root комментарии, без parentID)
func (r *CommentRepository) GetCommentsByPostBatch(
	ctx context.Context,
	postIDs []string,
) (map[string][]*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]*model.Comment)

	for _, postID := range postIDs {
		all := r.byPostID[postID]

		// фильтруем только root (parentID == nil)
		var roots []*model.Comment
		for _, c := range all {
			if c.ParentID == nil {
				roots = append(roots, c)
			}
		}

		result[postID] = roots
		if roots == nil {
			result[postID] = []*model.Comment{} // nil -> empty slice
		}
	}

	logrus.WithField("count", len(postIDs)).Debug("batch loaded comments by post from memory")
	return result, nil
}

// GetChildrenBatch получает дочерние комментарии для множества родителей
// Возвращает map[parentID] -> []*Comment
func (r *CommentRepository) GetChildrenBatch(
	ctx context.Context,
	parentIDs []string,
) (map[string][]*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]*model.Comment)

	for _, parentID := range parentIDs {
		children := r.byParentID[parentID]
		result[parentID] = children
		if children == nil {
			result[parentID] = []*model.Comment{} // nil -> empty slice
		}
	}

	logrus.WithField("count", len(parentIDs)).Debug("batch loaded children from memory")
	return result, nil
}