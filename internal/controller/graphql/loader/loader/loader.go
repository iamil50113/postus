package loader

import (
	"context"
	"net/http"
	"time"

	"github.com/vikstrous/dataloadgen"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
)

type CommentService interface {
	MultiChildExist(ctx context.Context, commentIDs []int64) ([]bool, []error)
}

type commentReader struct {
	commentStorage CommentProvider
}

type CommentProvider interface {
	MultiChildExist(ctx context.Context, commentIDs []*ComentAndPostID) ([]bool, []error)
}

// getUsers retrieves multiple users at the same time from the underlying storage system.
func (u commentReader) getChildCommentsExists(ctx context.Context, commentIDs []*ComentAndPostID) ([]bool, []error) {
	existsResults, err := u.commentStorage.MultiChildExist(ctx, commentIDs)
	return existsResults, err
}

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	CommentLoader *dataloadgen.Loader[*ComentAndPostID, bool]
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(s CommentProvider) *Loaders {
	// define the data loader
	cr := &commentReader{commentStorage: s}
	return &Loaders{
		CommentLoader: dataloadgen.NewLoader(cr.getChildCommentsExists, dataloadgen.WithWait(time.Millisecond*20)),
	}
}

// Middleware injects data loaders into the context
func Middleware(commentStorage CommentProvider, next http.Handler) http.Handler {
	// return a middleware that injects the loader to the request context
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Note that the loaders are being created per-request. This is important because they contain caching and batching logic that must be request-scoped.
		loaders := NewLoaders(commentStorage)
		r = r.WithContext(context.WithValue(r.Context(), loadersKey, loaders))
		next.ServeHTTP(w, r)
	})
}

// For returns the dataloader for a given context
func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

// GetChildCommentExists returns single user by id efficiently
func GetChildCommentExists(ctx context.Context, commentID int64, postID int64) (bool, error) {
	loaders := For(ctx)
	return loaders.CommentLoader.Load(ctx, &ComentAndPostID{CommentID: commentID, PostID: postID})
}

type ComentAndPostID struct {
	CommentID int64
	PostID    int64
}
