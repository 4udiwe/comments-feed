package app

import (
	"github.com/4udiwe/commnets-feed/config"
	comment_service "github.com/4udiwe/commnets-feed/internal/service/comment"
	post_service "github.com/4udiwe/commnets-feed/internal/service/post"
	"github.com/4udiwe/commnets-feed/pkg/transactor"
)

func (app *App) initServices() {
	app.postService = post_service.NewPostService(app.postRepo)

	txManager := app.getTransactor()

	app.commentService = comment_service.NewCommentService(
		app.commentRepo,
		app.commentBroker,
		app.postRepo,
		txManager,
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
