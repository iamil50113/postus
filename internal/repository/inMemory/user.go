package inmemory

import (
	"context"
	"postus/internal/domain/model"
	"postus/internal/repository"
)

func (s *Storage) User(ctx context.Context, uid int64) (*model.User, error) {
	const op = "repository.internalStorage.user.User"
	if uid >= int64(len(s.users)) {
		return nil, repository.ErrorUserNotFound
	}
	return &model.User{
		ID:   uid,
		Name: s.users[uid],
	}, nil
}