package comment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"postus/internal/domain/model"
	"postus/internal/repository"
	"postus/internal/service"
	"time"
	"unicode/utf8"
)

type Comment struct {
	log             *slog.Logger
	lenLimit        int
	paginationLimit int
	commProvider    CommentProvider
	commSaver       CommentSaver
	postProvider    PostProvider
	usrProvider     UserProvider
	subscriber      Subscriber
}

type CommentProvider interface {
	ChildExist(ctx context.Context, commentID int64) (bool, error)
	ChildCommentsForParentCommentIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error)
	CommentsForPostIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error)
	Comment(ctx context.Context, id int64) (*model.Comment, error)
}

type CommentSaver interface {
	NewComment(ctx context.Context, uid int64, postID int64, body string, publicationTime time.Time) (int64, error)
	NewChildComment(ctx context.Context, uid int64, postID int64, body string, parentCommentID int64, publicationTime time.Time) (int64, error)
}

type PostProvider interface {
	Post(ctx context.Context, id int64) (*model.Post, error)
}

type UserProvider interface {
	User(ctx context.Context, id int64) (*model.User, error)
}

func New(
	log *slog.Logger,
	lenLimit int,
	pagLimit int,
	commProvider CommentProvider,
	commSaver CommentSaver,
	postProvider PostProvider,
	usrProvider UserProvider) *Comment {
	return &Comment{
		log:             log,
		lenLimit:        lenLimit,
		paginationLimit: pagLimit,
		commProvider:    commProvider,
		commSaver:       commSaver,
		postProvider:    postProvider,
		usrProvider:     usrProvider,
		subscriber:      Subscriber{subs: make(map[int64][]Ch)},
	}
}

func (c *Comment) GetSubscriberService() *Subscriber {
	return &c.subscriber
}

func (c *Comment) ChildExist(ctx context.Context, commentID int64) (bool, error) {
	flag, err := c.commProvider.ChildExist(ctx, commentID)
	if err != nil {
		return false, service.ErrorServer
	}

	return flag, nil
}

func (c *Comment) ChildComments(ctx context.Context, id int64, cursor int64) (*model.Comments, error) {
	if _, err := c.commProvider.Comment(ctx, id); err != nil {
		if errors.Is(err, repository.ErrorCommentNotFound) {
			return nil, err
		} else {
			return nil, service.ErrorServer
		}
	}

	comments, err := c.commProvider.ChildCommentsForParentCommentIDWithCursor(ctx, id, cursor, c.paginationLimit)
	if err != nil {
		return nil, service.ErrorServer
	}

	return comments, nil
}

func (c *Comment) Comments(ctx context.Context, id int64, cursor int64) (*model.Comments, error) {
	if _, err := c.postProvider.Post(ctx, id); err != nil {
		if errors.Is(err, repository.ErrorPostNotFound) {
			return nil, err
		} else {
			return nil, service.ErrorServer
		}
	}

	comments, err := c.commProvider.CommentsForPostIDWithCursor(ctx, id, cursor, c.paginationLimit)
	if err != nil {
		return nil, service.ErrorServer
	}

	return comments, nil
}

func (c *Comment) NewComment(ctx context.Context, uid int64, postID int64, body string, parentCommentID int64) (int64, error) {
	if utf8.RuneCountInString(body) > c.lenLimit {
		return 0, fmt.Errorf("Comment length exceeded")
	}

	user, err := c.usrProvider.User(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrorUserNotFound) {
			return 0, err
		} else {
			return 0, service.ErrorServer
		}
	}

	post, err := c.postProvider.Post(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrorPostNotFound) {
			return 0, err
		} else {
			return 0, service.ErrorServer
		}
	}

	if !post.CommentPermission {
		return 0, fmt.Errorf("comments are disabled on this post")
	}

	publicationTime := time.Now()

	var newCommentID int64

	if parentCommentID == 0 {
		newCommentID, err = c.commSaver.NewComment(ctx, uid, postID, body, publicationTime)
		if err != nil {
			return 0, service.ErrorServer
		}
	} else {
		if _, err := c.commProvider.Comment(ctx, parentCommentID); err != nil {
			if errors.Is(err, repository.ErrorCommentNotFound) {
				return 0, err
			} else {
				return 0, service.ErrorServer
			}
		}
		newCommentID, err = c.commSaver.NewChildComment(ctx, uid, postID, body, parentCommentID, publicationTime)
		if err != nil {
			return 0, service.ErrorServer
		}
	}

	c.subscriber.NewPostAlert(&model.Comment{newCommentID, body, *user, postID, publicationTime})

	return newCommentID, nil
}
