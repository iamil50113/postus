package post

import (
	"context"
	"errors"
	"log/slog"
	"postus/internal/domain/model"
	"postus/internal/repository"
	"postus/internal/service"
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

func (p *Post) Posts(ctx context.Context, cursorID int64, limit *int64) ([]*model.Post, error) {
	posts, err := p.postProvider.Posts(ctx)
	if err != nil {
		return nil, service.ErrorServer
	}

	return posts, nil
}

func (p *Post) PostsForUser(ctx context.Context, uid int64) ([]*model.Post, error) {
	_, err := p.usrProvider.User(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrorUserNotFound) {
			return nil, err
		} else {
			return nil, service.ErrorServer
		}
	}

	return p.postProvider.PostsForUserID(ctx, uid)
}

func (p *Post) Post(ctx context.Context, id int64) (*model.Post, error) {
	post, err := p.postProvider.Post(ctx, id)

	if err != nil {
		if errors.Is(err, repository.ErrorPostNotFound) {
			return nil, err
		} else {
			return nil, service.ErrorServer
		}
	}

	return post, nil
}

func (p *Post) AddPost(ctx context.Context, userID int64, title string, body string, commentPermission bool) (int64, error) {
	_, err := p.usrProvider.User(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrorUserNotFound) {
			return 0, err
		} else {
			return 0, service.ErrorServer
		}
	}

	id, err := p.postSaver.NewPost(ctx, userID, title, body, commentPermission, time.Now())
	if err != nil {
		return 0, service.ErrorServer
	}

	return id, nil
}
