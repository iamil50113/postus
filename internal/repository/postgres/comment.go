package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"postus/internal/controller/graphql/loader/loader"
	"postus/internal/domain/model"
	"postus/internal/repository"
	"strconv"
	"time"
)

func (s *Storage) MultiChildExist(ctx context.Context, commentIDs []*loader.ComentAndPostID, postID int64) ([]bool, []error, error) {
	const op = "repository.postgres.comment.MultiChildExist"
	println(op)

	params := make([]interface{}, len(commentIDs))
	for i, v := range commentIDs {
		params[i] = v.CommentID
	}

	var paramrefs string

	for i, _ := range params {
		paramrefs += `$` + strconv.Itoa(i+1) + `,`
	}
	paramrefs = paramrefs[:len(paramrefs)-1] // remove last ","

	query := `SELECT c1.id, EXISTS(SELECT comment.id FROM comment WHERE comment.post_id = ` + strconv.Itoa(int(postID)) + ` AND parent_comment_id = c1.id)
				FROM comment c1
				WHERE c1.id IN (` + paramrefs + `)`

	rows, err := s.db.Query(ctx, query, params...)

	if err != nil {
		return nil, nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	defer rows.Close()

	errs := make([]error, 0, len(commentIDs))

	existsMap := make(map[int64]bool)
	for rows.Next() {
		var exist bool
		var id int64

		err := rows.Scan(&id, &exist)

		existsMap[id] = exist

		errs = append(errs, err)
	}

	exists := make([]bool, len(commentIDs), len(commentIDs))
	for i, v := range commentIDs {
		exists[i] = existsMap[v.CommentID]
	}

	println("вышли:", op)
	return exists, errs, nil
}

//func (s *Storage) ChildExist(ctx context.Context, commentID int64) (bool, error) {
//
//	const op = "repository.postgres.comment.ChildExist"
//	query := `SELECT EXISTS (SELECT * FROM comment WHERE parent_comment_id = $1);`
//
//	var exists bool
//	err := s.db.QueryRow(ctx, query, commentID).Scan(&exists)
//	if err != nil {
//		return false, fmt.Errorf("%s: execute statement: %w", op, err)
//	}
//	return exists, nil
//}

func (s *Storage) MultiFirstChildComments(ctx context.Context, commentIDs []int64, limit int) ([]*model.Comments, []error, error) {
	const op = "repository.postgres.comment.MultiFirstChildComments"
	println(op)

	params := make([]interface{}, len(commentIDs))
	for i, v := range commentIDs {
		params[i] = v
	}

	var paramrefs string

	for i, _ := range params {
		paramrefs += `$` + strconv.Itoa(i+1) + `,`
	}
	paramrefs = paramrefs[:len(paramrefs)-1] // remove last ","

	query := `SELECT *
FROM (
  SELECT comment.id, parent_comment_id, body, post_id, publication_time, user_id, name, row_number() OVER (PARTITION BY parent_comment_id ORDER BY comment.id) AS n
  FROM comment
	JOIN users ON comment.user_id = users.id
  WHERE parent_comment_id IN (` + paramrefs + `)
) x
WHERE n < 7
ORDER BY id;`

	rows, err := s.db.Query(ctx, query, params...)

	if err != nil {
		return nil, nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	defer rows.Close()

	errs := make([]error, 0, len(commentIDs))

	results := make([]*model.Comments, 0, len(commentIDs))

	for range commentIDs {
		results = append(results, &model.Comments{Comments: make([]model.Comment, 0, limit+1)})
		errs = append(errs, nil)
	}
	for rows.Next() {
		c := model.Comment{}
		var parentCommentId int64
		var n int

		err := rows.Scan(&c.ID, &parentCommentId, &c.Body, &c.PostID, &c.PublicationTime, &c.User.ID, &c.User.Name, &n)

		for k, v := range commentIDs {
			if v == parentCommentId {
				results[k].Comments = append(results[k].Comments, c)
				errs[k] = err
			}
		}
	}

	for _, v := range results {
		if len(v.Comments) > limit {
			v.HiddenComments = true
			v.Comments = v.Comments[:limit]
		}
	}
	println("вышли:", op)
	return results, errs, nil
}

func (s *Storage) ChildCommentsForParentCommentIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	const op = "repository.postgres.comment.ChildCommentsForParentCommentIDWithCursor"
	println(op)
	query := `SELECT comment.id, body, post_id, publication_time, user_id, name
				FROM comment
    			JOIN users ON comment.user_id = users.id
				WHERE comment.id > $1 AND parent_comment_id = $2
				ORDER BY comment.id
				LIMIT $3`
	rows, err := s.db.Query(ctx, query, cursor, id, limit+1)
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	cms := make([]model.Comment, 0, 6)
	for rows.Next() {
		c := model.Comment{}
		err := rows.Scan(&c.ID, &c.Body, &c.PostID, &c.PublicationTime, &c.User.ID, &c.User.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: execute statement: %w", op, err)
		}
		cms = append(cms, c)
	}

	if len(cms) > limit {
		return &model.Comments{Comments: cms[:limit], HiddenComments: true}, nil
	}

	return &model.Comments{Comments: cms, HiddenComments: false}, nil
}

func (s *Storage) CommentsForPostIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	const op = "repository.postgres.comment.GetCommentsForPostIDWithCursor"
	println(op)
	query := `SELECT comment.id, body, publication_time, user_id, name
				FROM comment
    			JOIN users ON comment.user_id = users.id
                WHERE comment.id > $1 AND post_id = $2 AND parent_comment_id IS NULL
                ORDER BY comment.id
                LIMIT $3`
	rows, err := s.db.Query(ctx, query, cursor, id, limit+1)
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	cms := make([]model.Comment, 0, 6)
	for rows.Next() {
		c := model.Comment{}
		err := rows.Scan(&c.ID, &c.Body, &c.PublicationTime, &c.User.ID, &c.User.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: execute statement: %w", op, err)
		}
		cms = append(cms, c)
	}
	if len(cms) > limit {
		return &model.Comments{Comments: cms[:limit], HiddenComments: true}, nil
	}

	return &model.Comments{Comments: cms, HiddenComments: false}, nil
}

func (s *Storage) NewComment(ctx context.Context, uid int64, postID int64, body string, publicationTime time.Time) (int64, error) {
	const op = "repository.postgres.comment.NewComment"
	println(op)
	query := `INSERT INTO comment (body, user_id, post_id, publication_time)
VALUES (@body, @userID, @postID, @publicationTime) RETURNING id`

	args := pgx.NamedArgs{
		"body":            body,
		"userID":          uid,
		"postID":          postID,
		"publicationTime": publicationTime,
	}

	var id int64
	err := s.db.QueryRow(ctx, query, args).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return id, nil
}

func (s *Storage) NewChildComment(ctx context.Context, uid int64, postID int64, body string, parentCommentID int64, publicationTime time.Time) (int64, error) {
	const op = "repository.postgres.comment.NewComment"
	println(op)
	query := `INSERT INTO comment (body, user_id, post_id, publication_time, parent_comment_id)
VALUES (@body, @userID, @postID, @publicationTime, @parentCommentID) RETURNING id`

	args := pgx.NamedArgs{
		"body":            body,
		"userID":          uid,
		"postID":          postID,
		"publicationTime": publicationTime,
		"parentCommentID": parentCommentID,
	}

	var id int64
	err := s.db.QueryRow(ctx, query, args).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return id, nil
}

func (s *Storage) Comment(ctx context.Context, id int64) (*model.Comment, error) {
	const op = "repository.postgres.comment.Comment"
	println(op)
	query := `SELECT body, user_id, post_id, publication_time from comment WHERE id = $1`
	var c model.Comment
	err := s.db.QueryRow(ctx, query, id).Scan(&c.Body, &c.User.ID, &c.PostID, &c.PublicationTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrorCommentNotFound
		}
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return &c, nil
}
