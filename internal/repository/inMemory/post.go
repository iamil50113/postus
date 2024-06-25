package inmemory

import (
	"golang.org/x/net/context"
	"postus/internal/domain/model"
	"postus/internal/repository"
	inmemorymodel "postus/internal/repository/inMemory/model"
	"time"
)

func (s *Storage) NewPost(ctx context.Context, userID int64, title string, body string, commentPermission bool, publicationTime time.Time) (int64, error) {
	const op = "repository.internalStorage.post.NewPost"
	id := int64(len(s.posts))

	s.posts = append(s.posts, &inmemorymodel.Post{
		Title:             title,
		Body:              body,
		UserID:            userID,
		PublicationTime:   publicationTime,
		CommentPermission: commentPermission,
	})

	return id, nil
}

func (s *Storage) Posts(ctx context.Context) ([]*model.Post, error) {
	const op = "repository.internalStorage.post.Posts"
	posts := make([]*model.Post, 0, 10)
	for id, v := range s.posts {
		posts = append(posts, &model.Post{
			ID:                int64(id),
			Title:             v.Title,
			Body:              v.Body,
			User:              model.User{ID: v.UserID, Name: s.users[v.UserID]},
			PublicationTime:   v.PublicationTime,
			CommentPermission: v.CommentPermission})
	}
	return posts, nil
}

func (s *Storage) PostsForUserID(ctx context.Context, uid int64) ([]*model.Post, error) {
	const op = "repository.internalStorage.post.PostsForUserID"
	posts := make([]*model.Post, 0, 10)
	for id, v := range s.posts {
		if v.UserID == uid {
			posts = append(posts, &model.Post{
				ID:                int64(id),
				Title:             v.Title,
				Body:              v.Body,
				User:              model.User{ID: v.UserID, Name: s.users[v.UserID]},
				PublicationTime:   v.PublicationTime,
				CommentPermission: v.CommentPermission})
		}
	}
	return posts, nil
}

func (s *Storage) Post(ctx context.Context, id int64) (*model.Post, error) {
	const op = "repository.internalStorage.post.Post"
	if id >= int64(len(s.posts)) {
		return nil, repository.ErrorPostNotFound
	}

	return &model.Post{
		ID:    id,
		Title: s.posts[id].Title,
		Body:  s.posts[id].Body,
		User: model.User{ID: s.posts[id].UserID,
			Name: s.users[s.posts[id].UserID],
		},
		PublicationTime:   s.posts[id].PublicationTime,
		CommentPermission: s.posts[id].CommentPermission,
	}, nil
}
