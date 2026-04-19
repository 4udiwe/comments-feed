package app

import (
	"github.com/4udiwe/comments-feed/config"
	"github.com/4udiwe/comments-feed/internal/graph/errors"
	comment_service "github.com/4udiwe/comments-feed/internal/service/comment"
	post_service "github.com/4udiwe/comments-feed/internal/service/post"
	"github.com/4udiwe/comments-feed/pkg/transactor"
)

func (app *App) initServices() {
	app.postService = post_service.NewPostService(
		app.postRepo,
		*errors.NewInputValidator(),
	)

	txManager := app.getTransactor()

	app.commentService = comment_service.NewCommentService(
		app.commentRepo,
		app.commentBroker,
		app.postRepo,
		txManager,
		*errors.NewInputValidator(),
	)
}

func (app *App) getTransactor() transactor.Transactor {
	switch app.cfg.Storage.Type {
	case config.StorageMemory:
		return transactor.NewNoOpTransactor()
	case config.StoragePostgres:
		return app.postgres
	default:
		return transactor.NewNoOpTransactor()
	}
}
