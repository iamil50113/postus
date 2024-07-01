package loader2

import (
	"context"
	"net/http"
	"time"

	"github.com/vikstrous/dataloadgen"
	"postus/internal/domain/model"
)

//	type CommentService interface {
//		MultiChildExist(ctx context.Context, commentIDs []int64) ([]bool, []error)
//		MultiFirstChildComments(ctx context.Context, commentIDs []int64) ([]*model.Comments, error)
//	}

type ctxKey string

const (
	loadersKey2 = ctxKey("dataloaders2")
)

type commentReader2 struct {
	childCommentsStorage CommentProvider2
}

type CommentProvider2 interface {
	MultiFirstChildComments(ctx context.Context, commentIDs []int64) ([]*model.Comments, []error)
}

func (u commentReader2) getFirstChildComments(ctx context.Context, commentIDs []int64) ([]*model.Comments, []error) {
	return u.childCommentsStorage.MultiFirstChildComments(ctx, commentIDs)
}

type Loaders2 struct {
	FirstCommentsLoader2 *dataloadgen.Loader[int64, *model.Comments]
}

func NewLoaders2(s CommentProvider2) *Loaders2 {
	// define the data loader
	cr := &commentReader2{childCommentsStorage: s}
	return &Loaders2{
		FirstCommentsLoader2: dataloadgen.NewLoader(cr.getFirstChildComments, dataloadgen.WithWait(time.Millisecond*20)),
	}
}

func Middleware2(commentStorage CommentProvider2, next http.Handler) http.Handler {
	// return a middleware that injects the loader to the request context
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Note that the loaders are being created per-request. This is important because they contain caching and batching logic that must be request-scoped.
		loaders := NewLoaders2(commentStorage)
		r = r.WithContext(context.WithValue(r.Context(), loadersKey2, loaders))
		next.ServeHTTP(w, r)
	})
}

func For2(ctx context.Context) *Loaders2 {
	return ctx.Value(loadersKey2).(*Loaders2)
}

func GetFirstChildComments(ctx context.Context, commentID int64) (*model.Comments, error) {
	loaders2 := For2(ctx)
	return loaders2.FirstCommentsLoader2.Load(ctx, commentID)
}
