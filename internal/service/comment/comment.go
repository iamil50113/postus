package comment

import (
	"context"
	"fmt"
	"log/slog"
	"postus/internal/domain/model"
	"time"
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
	return c.commProvider.ChildExist(ctx, commentID)
}

func (c *Comment) ChildComments(ctx context.Context, id int64, cursor int64) (*model.Comments, error) {
	return c.commProvider.ChildCommentsForParentCommentIDWithCursor(ctx, id, cursor, c.paginationLimit)
}

func (c *Comment) Comments(ctx context.Context, id int64, cursor int64) (*model.Comments, error) {
	if _, err := c.postProvider.Post(ctx, id); err != nil {
		return nil, err
	}
	return c.commProvider.CommentsForPostIDWithCursor(ctx, id, cursor, c.paginationLimit)
}

func (c *Comment) NewComment(ctx context.Context, uid int64, postID int64, body string, parentCommentID int64) (int64, error) {
	if len(body) > c.lenLimit {
		return 0, fmt.Errorf("Comment length exceeded")
	}

	user, err := c.usrProvider.User(ctx, uid)
	if err != nil {
		return 0, err
	}

	post, err := c.postProvider.Post(ctx, postID)
	if err != nil {
		return 0, err
	}

	if !post.CommentPermission {
		return 0, fmt.Errorf("comments are disabled on this post")
	}

	publicationTime := time.Now()

	var id int64

	if parentCommentID == 0 {
		id, err = c.commSaver.NewComment(ctx, uid, postID, body, publicationTime)
		if err != nil {
			return 0, err
		}
	} else {
		if _, err := c.commProvider.Comment(ctx, parentCommentID); err != nil {
			return 0, err
		}
		id, err = c.commSaver.NewChildComment(ctx, uid, postID, body, parentCommentID, publicationTime)
		if err != nil {
			return 0, err
		}
	}

	c.subscriber.NewPostAlert(&model.Comment{id, body, *user, postID, publicationTime})

	return id, nil
}
