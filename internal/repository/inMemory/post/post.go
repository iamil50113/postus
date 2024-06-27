package inmemoryPost

import (
	"golang.org/x/net/context"
	"postus/internal/domain/model"
	"postus/internal/repository"
	inmemorymodel "postus/internal/repository/inMemory/model"
	"sync"
	"time"
)

type DataStore struct {
	sync.RWMutex
	storage     []*inmemorymodel.Post
	usrProvider UserProvider
}

type UserProvider interface {
	User(ctx context.Context, uid int64) (*model.User, error)
}

func New(usrProvider UserProvider) *DataStore {
	return &DataStore{
		storage:     make([]*inmemorymodel.Post, 1, 10),
		usrProvider: usrProvider,
	}
}

func (s *DataStore) NewPost(ctx context.Context, userID int64, title string, body string, commentPermission bool, publicationTime time.Time) (int64, error) {
	s.Lock()
	defer s.Unlock()
	return s.newPost(ctx, userID, title, body, commentPermission, publicationTime)
}

func (s *DataStore) Posts(ctx context.Context) ([]*model.Post, error) {
	s.RLock()
	defer s.RUnlock()
	return s.posts(ctx)
}

func (s *DataStore) PostsForUserID(ctx context.Context, uid int64) ([]*model.Post, error) {
	s.RLock()
	defer s.RUnlock()
	return s.postsForUserID(ctx, uid)
}

func (s *DataStore) Post(ctx context.Context, id int64) (*model.Post, error) {
	s.RLock()
	defer s.RUnlock()
	return s.post(ctx, id)
}

func (s *DataStore) newPost(ctx context.Context, userID int64, title string, body string, commentPermission bool, publicationTime time.Time) (int64, error) {
	id := int64(len(s.storage))

	s.storage = append(s.storage, &inmemorymodel.Post{
		Title:             title,
		Body:              body,
		UserID:            userID,
		PublicationTime:   publicationTime,
		CommentPermission: commentPermission,
	})

	return id, nil
}

func (s *DataStore) posts(ctx context.Context) ([]*model.Post, error) {
	posts := make([]*model.Post, 0, 10)

	for id, v := range s.storage {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		default:
			user, err := s.usrProvider.User(ctx, v.UserID)
			if err != nil {
				continue
			}

			posts = append(posts, &model.Post{
				ID:    int64(id),
				Title: v.Title,
				Body:  v.Body,
				User: model.User{
					ID:   user.ID,
					Name: user.Name,
				},
				PublicationTime:   v.PublicationTime,
				CommentPermission: v.CommentPermission})
		}
	}
	return posts, nil
}

func (s *DataStore) postsForUserID(ctx context.Context, uid int64) ([]*model.Post, error) {
	posts := make([]*model.Post, 0, 10)

	for id, v := range s.storage {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		default:

			if v.UserID == uid {
				user, err := s.usrProvider.User(ctx, v.UserID)
				if err != nil {
					continue
				}
				posts = append(posts, &model.Post{
					ID:    int64(id),
					Title: v.Title,
					Body:  v.Body,
					User: model.User{
						ID:   user.ID,
						Name: user.Name,
					},
					PublicationTime:   v.PublicationTime,
					CommentPermission: v.CommentPermission})
			}
		}
	}
	return posts, nil
}

func (s *DataStore) post(ctx context.Context, id int64) (*model.Post, error) {
	if id >= int64(len(s.storage)) {
		return nil, repository.ErrorPostNotFound
	}

	user, err := s.usrProvider.User(ctx, s.storage[id].UserID)
	if err != nil {
		return nil, repository.ErrorUserNotFound
	}

	return &model.Post{
		ID:    id,
		Title: s.storage[id].Title,
		Body:  s.storage[id].Body,
		User: model.User{
			ID:   user.ID,
			Name: user.Name,
		},
		PublicationTime:   s.storage[id].PublicationTime,
		CommentPermission: s.storage[id].CommentPermission,
	}, nil
}
