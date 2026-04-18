package app

import (
	"log"
	"os"

	"github.com/4udiwe/commnets-feed/config"
	"github.com/4udiwe/commnets-feed/internal/broker"
	"github.com/4udiwe/commnets-feed/internal/graph"
	"github.com/4udiwe/commnets-feed/internal/repository"
	comment_service "github.com/4udiwe/commnets-feed/internal/service/comment"
	post_service "github.com/4udiwe/commnets-feed/internal/service/post"
	"github.com/4udiwe/commnets-feed/pkg/postgres"
	"github.com/99designs/gqlgen/graphql/handler"
)

type App struct {
	cfg       *config.Config
	interrupt <-chan os.Signal

	// DB
	postgres *postgres.Postgres

	// Repositories
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository

	// Broker
	commentBroker *broker.CommentBroker

	// Services
	postService    *post_service.PostService
	commentService *comment_service.CommentService
}

func New(configPath string) *App {
	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatalf("app - New - config.New: %v", err)
	}

	initLogger(cfg.Log.Level)

	return &App{
		cfg: cfg,
	}
}

func (app *App) Start() {
	app.initRepositories()
	app.commentBroker = broker.NewCommentBroker()
	app.initServices()

	srv := handler.New(graph.NewExecutableSchema(
		graph.Config{
			Resolvers: &graph.Resolver{
				PostService:    *app.postService,
				CommentService: *app.commentService,
			},
		},
	))

	app.configureGraphQL(srv)
	app.runHTTP(srv)
}


func (app *App) Stop() {
	log.Println("stopping application...")

	if app.postgres != nil {
		app.postgres.Close()
	}
}