package graph

import (
	"context"

	"github.com/4udiwe/comments-feed/internal/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type PostService interface {
	CreatePost(ctx context.Context, input model.CreatePostInput) (*model.Post, error)
	GetPost(ctx context.Context, id string) (*model.Post, error)
	ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error)
}

type CommentService interface {
	CreateComment(ctx context.Context, input model.CreateCommentInput) (*model.Comment, error)
	GetCommentsByPost(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error)
	//GetChildren(ctx context.Context, parentID string, limit, offset int) ([]*model.Comment, error)
	SubscribeToComments(ctx context.Context, postID string) (<-chan *model.Comment, error)
	//GetCommentsByPostWithPagination(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error)
	//GetChildrenWithPagination(ctx context.Context, parentID string, limit, offset int) ([]*model.Comment, error)
}

type Resolver struct {
	PostService    PostService
	CommentService CommentService
}
