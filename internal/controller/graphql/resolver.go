package graph

import (
	"context"
	"postus/internal/domain/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	postService PostService
	commService CommentService
	usrService  UserService
	subService  SubscriberService
}

type PostService interface {
	Posts(ctx context.Context, cursorID int64, limit *int64) ([]*model.Post, error)
	PostsForUser(ctx context.Context, uid int64) ([]*model.Post, error)
	Post(ctx context.Context, id int64) (*model.Post, error)
	AddPost(ctx context.Context, userID int64, title string, body string, commentPermission bool) (int64, error)
}

type CommentService interface {
	ChildComments(ctx context.Context, id int64, cursor int64) (*model.Comments, error)
	Comments(ctx context.Context, id int64, cursor int64) (*model.Comments, error)
	NewComment(ctx context.Context, uid int64, postID int64, body string, parentCommentID *int64) (int64, error)
	CommentsForUser(ctx context.Context, uid int64) (*model.Comments, error)
}

type SubscriberService interface {
	NewSubscribe(ctx context.Context, postID int64) (<-chan *model.Comment, error)
}

type UserService interface {
	User(ctx context.Context, uid int64) (*model.User, error)
}

func New(post PostService, comm CommentService, usr UserService, sub SubscriberService) *Resolver {
	return &Resolver{postService: post, commService: comm, usrService: usr, subService: sub}
}
