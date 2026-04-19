package app

import (
	"log"

	"github.com/4udiwe/comments-feed/config"
	memory_comment_repository "github.com/4udiwe/comments-feed/internal/repository/memory/comment"
	memory_post_repository "github.com/4udiwe/comments-feed/internal/repository/memory/post"
	postgres_comment_repository "github.com/4udiwe/comments-feed/internal/repository/postgres/comment"
	postgres_post_repository "github.com/4udiwe/comments-feed/internal/repository/postgres/post"
)

func (app *App) initRepositories() {
	switch app.cfg.Storage.Type {

	case config.StorageMemory:
		app.postRepo = memory_post_repository.NewPostRepository()
		app.commentRepo = memory_comment_repository.NewCommentRepository()

	case config.StoragePostgres:
		app.initPostgres()
		app.postRepo = postgres_post_repository.New(app.postgres)
		app.commentRepo = postgres_comment_repository.New(app.postgres)

	default:
		log.Fatalf("unknown storage type: %s", app.cfg.Storage.Type)
	}
}
