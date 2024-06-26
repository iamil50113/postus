package inmemoryComment

import (
	"context"
	"postus/internal/domain/model"
	"postus/internal/repository"
	inmemorymodel "postus/internal/repository/inMemory/model"
	"sync"
	"time"
)

type DataStore struct {
	sync.RWMutex
	storage     []*inmemorymodel.Comment
	usrProvider UserProvider
}

type UserProvider interface {
	User(ctx context.Context, uid int64) (*model.User, error)
}

func New(usrProvider UserProvider) *DataStore {
	return &DataStore{
		storage:     make([]*inmemorymodel.Comment, 10),
		usrProvider: usrProvider,
	}
}

func (s *DataStore) ChildExist(ctx context.Context, commentID int64) (bool, error) {
	s.RLock()
	defer s.RUnlock()

	for _, v := range s.storage {
		select {
		case <-ctx.Done():
			return false, ctx.Err()

		default:
			if v.ParentCommentID == commentID {
				return true, nil
			}
		}
	}

	return false, nil
}

func (s *DataStore) ChildCommentsForParentCommentIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	comms := []model.Comment{}

	s.RLock()
	defer s.RUnlock()

	for i := cursor; i < int64(len(s.storage)) && len(comms) < limit+1; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		default:
			if s.storage[i].ParentCommentID == id {
				user, err := s.usrProvider.User(ctx, s.storage[i].UserID)
				if err != nil {
					continue
				}

				comms = append(comms, model.Comment{
					ID:   i,
					Body: s.storage[i].Body,
					User: model.User{
						ID:   user.ID,
						Name: user.Name},
					PostID:          s.storage[i].PostID,
					PublicationTime: s.storage[i].PublicationTime,
				})
			}
		}
	}
	if len(comms) > limit {
		return &model.Comments{Comments: comms[:limit], HiddenComments: true}, nil
	}

	return &model.Comments{Comments: comms, HiddenComments: false}, nil
}

func (s *DataStore) CommentsForPostIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	comms := []model.Comment{}

	s.RLock()
	defer s.RUnlock()

	for i := cursor; i < int64(len(s.storage)) && len(comms) < limit+1; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		default:
			if s.storage[i].PostID == id && s.storage[i].ParentCommentID == 0 {
				user, err := s.usrProvider.User(ctx, s.storage[i].UserID)
				if err != nil {
					continue
				}

				comms = append(comms, model.Comment{
					ID:              int64(i),
					Body:            s.storage[i].Body,
					User:            model.User{ID: user.ID, Name: user.Name},
					PostID:          s.storage[i].PostID,
					PublicationTime: s.storage[i].PublicationTime,
				})
			}
		}
	}
	if len(comms) > limit {
		return &model.Comments{Comments: comms[:limit], HiddenComments: true}, nil
	}

	return &model.Comments{Comments: comms, HiddenComments: false}, nil

}

func (s *DataStore) NewComment(ctx context.Context, uid int64, postID int64, body string, publicationTime time.Time) (int64, error) {
	s.Lock()
	defer s.Unlock()

	id := int64(len(s.storage))

	s.storage = append(s.storage, &inmemorymodel.Comment{
		Body:            body,
		UserID:          uid,
		PostID:          postID,
		ParentCommentID: 0,
		PublicationTime: publicationTime})

	return id, nil
}

func (s *DataStore) NewChildComment(ctx context.Context, uid int64, postID int64, body string, parentCommentID int64, publicationTime time.Time) (int64, error) {
	s.Lock()
	defer s.Unlock()

	id := int64(len(s.storage))

	s.storage = append(s.storage, &inmemorymodel.Comment{
		Body:            body,
		UserID:          uid,
		PostID:          postID,
		ParentCommentID: parentCommentID,
		PublicationTime: publicationTime})

	return id, nil
}

func (s *DataStore) Comment(ctx context.Context, id int64) (*model.Comment, error) {
	s.RLock()
	defer s.RUnlock()

	if id >= int64(len(s.storage)) {
		return nil, repository.ErrorCommentNotFound
	}

	user, err := s.usrProvider.User(ctx, s.storage[id].UserID)
	if err != nil {
		return nil, repository.ErrorUserNotFound
	}

	return &model.Comment{
		ID:   id,
		Body: s.storage[id].Body,
		User: model.User{
			ID:   user.ID,
			Name: user.Name,
		},
		PostID:          s.storage[id].PostID,
		PublicationTime: s.storage[id].PublicationTime,
	}, nil
}
