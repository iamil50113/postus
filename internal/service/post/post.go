package post

import (
	"context"
	"fmt"
	"log/slog"
	"postus/internal/domain/model"
	"time"
)

type Post struct {
	log          *slog.Logger
	postProvider PostProvider
	postSaver    PostSaver
	usrProvider  UserProvider
}

type PostProvider interface {
	Posts(ctx context.Context) ([]*model.Post, error)
	PostsForUserID(ctx context.Context, uid int64) ([]*model.Post, error)
	Post(ctx context.Context, id int64) (*model.Post, error)
}

type PostSaver interface {
	NewPost(ctx context.Context, userID int64, title string, body string, commentPermission bool, publicationTime time.Time) (int64, error)
}

type UserProvider interface {
	User(ctx context.Context, id int64) (*model.User, error)
}

func New(
	log *slog.Logger,
	postProvider PostProvider,
	postSaver PostSaver,
	userProvider UserProvider) *Post {
	return &Post{
		log:          log,
		postProvider: postProvider,
		postSaver:    postSaver,
		usrProvider:  userProvider,
	}
}

func (p *Post) Posts(ctx context.Context) ([]*model.Post, error) {
	return p.postProvider.Posts(ctx)
}

func (p *Post) PostsForUser(ctx context.Context, uid int64) ([]*model.Post, error) {
	user, err := p.usrProvider.User(ctx, uid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("Invalid user id")
	}
	return p.postProvider.PostsForUserID(ctx, uid)
}

func (p *Post) Post(ctx context.Context, id int64) (*model.Post, error) {
	post, err := p.postProvider.Post(ctx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, fmt.Errorf("Invalid post id")
	}
	return post, err
}

func (p *Post) AddPost(ctx context.Context, userID int64, title string, body string, commentPermission bool) (int64, error) {
	_, err := p.usrProvider.User(ctx, userID)
	if err != nil {
		return 0, err
	}

	return p.postSaver.NewPost(ctx, userID, title, body, commentPermission, time.Now())
}
