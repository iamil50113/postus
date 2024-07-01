package inmemoryComment

import (
	"context"
	"postus/internal/controller/graphql/loader/loader"
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
		storage:     make([]*inmemorymodel.Comment, 1, 10),
		usrProvider: usrProvider,
	}
}

func (s *DataStore) MultiChildExist(ctx context.Context, commentIDs []*loader.ComentAndPostID, postID int64) ([]bool, []error, error) {
	s.RLock()
	defer s.RUnlock()
	return s.multiChildExist(ctx, commentIDs, postID)
}

//func (s *DataStore) ChildExist(ctx context.Context, commentID int64) (bool, error) {
//	s.RLock()
//	defer s.RUnlock()
//	return s.childExist(ctx, commentID)
//}

func (s *DataStore) MultiFirstChildComments(ctx context.Context, commentIDs []int64, limit int) ([]*model.Comments, []error, error) {
	return nil, nil, nil
}

func (s *DataStore) ChildCommentsForParentCommentIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	s.RLock()
	defer s.RUnlock()
	return s.childCommentsForParentCommentIDWithCursor(ctx, id, cursor, limit)
}

func (s *DataStore) CommentsForPostIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	s.RLock()
	defer s.RUnlock()
	return s.commentsForPostIDWithCursor(ctx, id, cursor, limit)
}

func (s *DataStore) NewComment(ctx context.Context, uid int64, postID int64, body string, publicationTime time.Time) (int64, error) {
	s.Lock()
	defer s.Unlock()
	return s.newComment(ctx, uid, postID, body, publicationTime)
}

func (s *DataStore) NewChildComment(ctx context.Context, uid int64, postID int64, body string, parentCommentID int64, publicationTime time.Time) (int64, error) {
	s.Lock()
	defer s.Unlock()
	return s.newChildComment(ctx, uid, postID, body, parentCommentID, publicationTime)
}

func (s *DataStore) Comment(ctx context.Context, id int64) (*model.Comment, error) {
	s.RLock()
	defer s.RUnlock()
	return s.comment(ctx, id)
}

func (s *DataStore) multiChildExist(ctx context.Context, commentIDs []*loader.ComentAndPostID, postID int64) ([]bool, []error, error) {
	exists := make([]bool, len(commentIDs), len(commentIDs))

commentIDs:
	for i, d := range commentIDs {
		for _, v := range s.storage {
			select {
			case <-ctx.Done():
				return nil, nil, ctx.Err()

			default:
				if v.ParentCommentID == d.CommentID {
					exists[i] = true
					continue commentIDs
				}
			}
		}
	}

	return exists, nil, nil
}

//func (s *DataStore) childExist(ctx context.Context, commentID int64) (bool, error) {
//	for _, v := range s.storage {
//		select {
//		case <-ctx.Done():
//			return false, ctx.Err()
//
//		default:
//			if v.ParentCommentID == commentID {
//				return true, nil
//			}
//		}
//	}
//
//	return false, nil
//}

func (s *DataStore) childCommentsForParentCommentIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	comms := []model.Comment{}

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

func (s *DataStore) commentsForPostIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	comms := []model.Comment{}

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

func (s *DataStore) newComment(ctx context.Context, uid int64, postID int64, body string, publicationTime time.Time) (int64, error) {
	id := int64(len(s.storage))

	s.storage = append(s.storage, &inmemorymodel.Comment{
		Body:            body,
		UserID:          uid,
		PostID:          postID,
		ParentCommentID: 0,
		PublicationTime: publicationTime})

	return id, nil
}

func (s *DataStore) newChildComment(ctx context.Context, uid int64, postID int64, body string, parentCommentID int64, publicationTime time.Time) (int64, error) {
	id := int64(len(s.storage))

	s.storage = append(s.storage, &inmemorymodel.Comment{
		Body:            body,
		UserID:          uid,
		PostID:          postID,
		ParentCommentID: parentCommentID,
		PublicationTime: publicationTime})

	return id, nil
}

func (s *DataStore) comment(ctx context.Context, id int64) (*model.Comment, error) {
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
