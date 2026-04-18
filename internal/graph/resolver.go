package graph

import (
	comment_service "github.com/4udiwe/commnets-feed/internal/service/comment"
	post_service "github.com/4udiwe/commnets-feed/internal/service/post"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	PostService    post_service.PostService
	CommentService comment_service.CommentService
}
