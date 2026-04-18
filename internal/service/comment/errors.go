package comment_service

import "errors"

var (
	ErrInvalidCommentLength  = errors.New("invalid comment length")
	ErrParentPostNotFound    = errors.New("parent post for comment not found")
	ErrDisabledComments      = errors.New("comments are disabled for this post")
	ErrParentCommentNotFound = errors.New("parent comment not found")
	ErrCommentFromOtherPost  = errors.New("parent comment belongs to another post")
	ErrCannotCreateComment   = errors.New("cannot create comment")
	ErrCannotFetchComments   = errors.New("cannot fetch comments")
)
