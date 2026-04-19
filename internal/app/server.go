package app

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	gqlmiddleware "github.com/4udiwe/comments-feed/internal/graph/middleware"
	httpmiddleware "github.com/4udiwe/comments-feed/internal/middleware"
	"github.com/4udiwe/comments-feed/pkg/httpserver"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

func (app *App) runHTTP(srv *handler.Server) {
	mux := http.NewServeMux()

	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)

	// DataLoader middleware
	// Он будет создавать новый набор DataLoaders для каждого HTTP запроса
	handler := gqlmiddleware.DataLoaderMiddleware(app.commentService)(mux)

	// HTTP logging middleware
	httpLoggingMiddleware := httpmiddleware.NewHTTPLoggingMiddleware()
	handler = httpLoggingMiddleware.Handler(handler)

	// Recover middleware
	recoverMiddleware := httpmiddleware.NewRecoverMiddleware()
	handler = recoverMiddleware(handler)

	httpServer := httpserver.New(
		handler,
		httpserver.Port(app.cfg.HTTP.Port),
	)

	httpServer.Start()

	log.Printf("server started on port %s, data store mode:%s", app.cfg.HTTP.Port, app.cfg.Storage.Type)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-httpServer.Notify():
		log.Fatalf("server error: %v", err)

	case <-stop:
		log.Println("shutting down server...")

		if err := httpServer.Shutdown(); err != nil {
			log.Fatalf("shutdown error: %v", err)
		}

		app.Stop()
	}
}
