package post_service

import "errors"

var (
	ErrPostNotFound       = errors.New("post not found")
	ErrPostAlreadyExists  = errors.New("post already exists")
	ErrCannotFetchPosts   = errors.New("cannot fetch posts")
	ErrCannotCreatePost   = errors.New("cannot create post")
	ErrInvalidPostContent = errors.New("invalid post content")
)
