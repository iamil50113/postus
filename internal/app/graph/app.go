package graphApp

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"postus/internal/config"
	graph "postus/internal/controller/graphql"
	"postus/internal/controller/graphql/loader/loader"
	"postus/internal/controller/graphql/loader/loader2"
	inmemory "postus/internal/repository/inMemory"
	"postus/internal/repository/postgres"
	"postus/internal/service/comment"
	"postus/internal/service/post"
	"postus/internal/service/user"
	"syscall"
	"time"
)

const (
	defaultPort = ":8080"
)

type App struct {
	log *slog.Logger
}

func New(log *slog.Logger) *App { return &App{log: log} }

func (a *App) MustRun(
	postusCfg config.PostusConfig,
	httpServerCfg config.HTTPServer,
	postgresCfg config.PostgresConfig,
) {

	if err := a.Run(postusCfg, httpServerCfg, postgresCfg); err != nil {
		panic(err)
	}
}

// Run runs GraphQL server.
func (a *App) Run(
	postusCfg config.PostusConfig,
	httpServerCfg config.HTTPServer,
	postgresCfg config.PostgresConfig,
) error {
	var postService *post.Post
	var commentService *comment.Comment
	var userService *user.User

	switch postusCfg.UseInMemory {
	case false:
		postgresPool, err := pgxpool.New(context.Background(), fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", postgresCfg.Host, postgresCfg.Port, postgresCfg.User, postgresCfg.Pass, postgresCfg.DBName))
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		defer postgresPool.Close()

		st, err := postgres.New(postgresPool)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		postService = post.New(a.log, st, st, st)
		commentService = comment.New(a.log, postusCfg.CommentLenLimit, postusCfg.PaginationCommentsLimit, st, st, st, st)
		userService = user.New(a.log, st)
	case true:
		st, err := inmemory.New()
		if err != nil {
			panic("inMemory storage initialization error")
		}

		postService = post.New(a.log, st.Posts, st.Posts, st.Users)
		commentService = comment.New(a.log, postusCfg.CommentLenLimit, postusCfg.PaginationCommentsLimit, st.Comments, st.Comments, st.Posts, st.Users)
		userService = user.New(a.log, st.Users)
	}

	graphResolver := graph.New(postService, commentService, userService, commentService.GetSubscriberService())

	port := httpServerCfg.Port
	if port == "" {
		port = defaultPort
	}

	graphConfig := graph.Config{Resolvers: graphResolver}

	countComplexity := func(childComplexity int, cursorID *int64, limit *int64) int {
		return childComplexity / 1000
	}
	graphConfig.Complexity.Query.Posts = countComplexity

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graphConfig))

	srv.Use(extension.FixedComplexityLimit(50))

	timer := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			h.ServeHTTP(w, r)
			a.log.Info("время обработки запроса: " + time.Now().Sub(startTime).String())
		})
	}
	serverWithTimer := timer(srv)
	serverWithLoaders := loader.Middleware(commentService, serverWithTimer)
	serverWithLoaders2 := loader2.Middleware2(commentService, serverWithLoaders)

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", serverWithLoaders2)

	serv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	a.log.Info(fmt.Sprintf("connect to http://localhost:%s/ for GraphQL playground", port))

	serverChan := make(chan error, 1)

	go func() {
		if err := serv.ListenAndServe(); err != nil {
			serverChan <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-stop:
		a.log.Info("stopping graphQL server")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := serv.Shutdown(ctx); err != nil {
			a.log.Error("server closed with err: %+v", err)
			os.Exit(1)
		}
		a.log.Info("Gracefully stopped")

	case err := <-serverChan:
		a.log.Error("server shutdown", err)
	}

	return nil
}
