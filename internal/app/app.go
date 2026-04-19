package app

import (
	"log"

	"github.com/4udiwe/comments-feed/config"
	"github.com/4udiwe/comments-feed/internal/broker"
	"github.com/4udiwe/comments-feed/internal/graph"
	comment_service "github.com/4udiwe/comments-feed/internal/service/comment"
	post_service "github.com/4udiwe/comments-feed/internal/service/post"
	"github.com/4udiwe/comments-feed/pkg/postgres"
	"github.com/99designs/gqlgen/graphql/handler"
)

type App struct {
	cfg *config.Config

	// DB
	postgres *postgres.Postgres

	// Repositories
	postRepo    PostRepository
	commentRepo CommentRepository

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
				PostService:    app.postService,
				CommentService: app.commentService,
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
