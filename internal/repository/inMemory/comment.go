package inmemory

import (
	"context"
	"postus/internal/domain/model"
	"postus/internal/repository"
	inmemorymodel "postus/internal/repository/inMemory/model"
	"time"
)

func (s *Storage) ChildExist(ctx context.Context, commentID int64) (bool, error) {
	for _, v := range s.comments {
		if v.ParentCommentID == commentID {
			return true, nil
		}
	}
	return false, nil
}

func (s *Storage) ChildCommentsForParentCommentIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	comms := []model.Comment{}
	for i := cursor; i < int64(len(s.comments)) && len(comms) < s.paginationLimit+1; i++ {
		if s.comments[i].ParentCommentID == id {
			comms = append(comms, model.Comment{
				ID:              int64(i),
				Body:            s.comments[i].Body,
				User:            model.User{ID: s.comments[i].UserID, Name: s.users[s.comments[i].UserID]},
				PostID:          s.comments[i].PostID,
				PublicationTime: s.comments[i].PublicationTime,
			})
		}
	}
	if len(comms) > limit {
		return &model.Comments{Comments: comms[:limit], HiddenComments: true}, nil
	}

	return &model.Comments{Comments: comms, HiddenComments: false}, nil
}

func (s *Storage) CommentsForPostIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	comms := []model.Comment{}
	for i := cursor; i < int64(len(s.comments)) && len(comms) < s.paginationLimit+1; i++ {
		if s.comments[i].PostID == id && s.comments[i].ParentCommentID == 0 {
			comms = append(comms, model.Comment{
				ID:              int64(i),
				Body:            s.comments[i].Body,
				User:            model.User{ID: s.comments[i].UserID, Name: s.users[s.comments[i].UserID]},
				PostID:          s.comments[i].PostID,
				PublicationTime: s.comments[i].PublicationTime,
			})
		}
	}
	if len(comms) > limit {
		return &model.Comments{Comments: comms[:limit], HiddenComments: true}, nil
	}

	return &model.Comments{Comments: comms, HiddenComments: false}, nil

}

func (s *Storage) NewComment(ctx context.Context, uid int64, postID int64, body string, publicationTime time.Time) (int64, error) {
	id := int64(len(s.comments))

	s.comments = append(s.comments, &inmemorymodel.Comment{
		Body:            body,
		UserID:          uid,
		PostID:          postID,
		ParentCommentID: 0,
		PublicationTime: publicationTime})

	return id, nil
}

func (s *Storage) NewChildComment(ctx context.Context, uid int64, postID int64, body string, parentCommentID int64, publicationTime time.Time) (int64, error) {
	id := int64(len(s.comments))

	s.comments = append(s.comments, &inmemorymodel.Comment{
		Body:            body,
		UserID:          uid,
		PostID:          postID,
		ParentCommentID: parentCommentID,
		PublicationTime: publicationTime})

	return id, nil
}

func (s *Storage) Comment(ctx context.Context, id int64) (*model.Comment, error) {
	if id >= int64(len(s.comments)) {
		return nil, repository.ErrorCommentNotFound
	}

	return &model.Comment{
		ID:   id,
		Body: s.comments[id].Body,
		User: model.User{ID: s.comments[id].UserID,
			Name: s.users[s.comments[id].UserID],
		},
		PostID:          s.comments[id].PostID,
		PublicationTime: s.comments[id].PublicationTime,
	}, nil
}
