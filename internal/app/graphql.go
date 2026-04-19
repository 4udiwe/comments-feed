package app

import (
	"context"
	"errors"
	"runtime/debug"
	"time"

	gqlerrors "github.com/4udiwe/comments-feed/internal/graph/errors"
	gqlmiddleware "github.com/4udiwe/comments-feed/internal/graph/middleware"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/v2/ast"
)

func (app *App) configureGraphQL(srv *handler.Server) {
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// custom error presenter для логирования и трансформации ошибок
	srv.SetErrorPresenter(gqlerrors.ErrorPresenter)

	// GraphQL operation logging middleware
	srv.AroundOperations(gqlmiddleware.OperationLoggingMiddleware)

	// resolver error logging middleware
	srv.AroundFields(gqlmiddleware.ResolverErrorLoggingMiddleware)

	// Ловим паники внутри resolver'ов
	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
		logrus.Errorf("panic in GraphQL: %v\n%s", err, debug.Stack())
		return errors.New("internal server error")
	})
}
