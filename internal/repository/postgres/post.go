package postgres

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
	"postus/internal/domain/model"
	"postus/internal/repository"
	"time"
)

func (s *Storage) NewPost(ctx context.Context, userID int64, title string, body string, commentPermission bool, publicationTime time.Time) (int64, error) {
	const op = "repository.postgres.post.NewPost"

	query := `INSERT INTO post (title, body, user_id, publication_time, comment_permission)
VALUES (@title, @body, @userID, @publicationTime, @commentPermission) RETURNING id`

	args := pgx.NamedArgs{
		"title":             title,
		"body":              body,
		"userID":            userID,
		"publicationTime":   publicationTime,
		"commentPermission": commentPermission,
	}

	var id int64
	err := s.db.QueryRow(ctx, query, args).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Storage) Posts(ctx context.Context) ([]*model.Post, error) {
	const op = "repository.postgres.post.Posts"

	query := `SELECT post.id, title, body, publication_time, comment_permission, user_id, name FROM post JOIN users ON post.user_id = users.id`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	posts := []*model.Post{}
	for rows.Next() {
		c := model.Post{}
		err := rows.Scan(&c.ID, &c.Title, &c.Body, &c.PublicationTime, &c.CommentPermission, &c.User.ID, &c.User.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: execute statement: %w", op, err)
		}
		posts = append(posts, &c)
	}

	return posts, nil
}

func (s *Storage) PostsForUserID(ctx context.Context, uid int64) ([]*model.Post, error) {
	const op = "repository.postgres.post.PostsForUserID"

	query := `SELECT post.id, title, body, name, publication_time, comment_permission
				FROM post
    			JOIN users ON post.user_id = users.id
                WHERE post.user_id = $1`

	rows, err := s.db.Query(ctx, query, uid)
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	posts := []*model.Post{}

	for rows.Next() {
		c := &model.Post{User: model.User{ID: uid}}
		err := rows.Scan(&c.ID, &c.Title, &c.Body, &c.User.Name, &c.PublicationTime, &c.CommentPermission)
		if err != nil {
			return nil, fmt.Errorf("%s: execute statement: %w", op, err)
		}
		posts = append(posts, c)
	}

	return posts, nil
}

func (s *Storage) Post(ctx context.Context, id int64) (*model.Post, error) {
	const op = "repository.postgres.post.Post"

	query := `SELECT title, body, publication_time, comment_permission, user_id, name FROM post JOIN users ON post.user_id = users.id WHERE post.id = $1`

	var p model.Post
	err := s.db.QueryRow(ctx, query, id).Scan(&p.Title, &p.Body, &p.PublicationTime, &p.CommentPermission, &p.User.ID, &p.User.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrorPostNotFound
		}
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return &p, nil

	//rows, err := s.db.Query(ctx, query, id)
	//if err != nil {
	//	return nil, err
	//}
	//defer rows.Close()
	//
	//var c *model.Post
	//for rows.Next() {
	//	err := rows.Scan(&c.Title, &c.Body, &c.PublicationTime, &c.CommentPermission, &c.User.ID, &c.User.Name)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	//return c, nil
}
