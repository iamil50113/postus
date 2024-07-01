package inmemoryUser

import (
	"context"
	"postus/internal/domain/model"
	"postus/internal/repository"
	"sync"
)

type DataStore struct {
	sync.RWMutex
	storage []string
}

func New() *DataStore {
	users := make([]string, 1, 100)
	users = append(users, "ivan")
	users = append(users, "egor")
	users = append(users, "alex")

	return &DataStore{
		storage: users,
	}
}

func (s *DataStore) User(ctx context.Context, uid int64) (*model.User, error) {
	s.RLock()
	defer s.RUnlock()
	return s.user(ctx, uid)
}

func (s *DataStore) user(ctx context.Context, uid int64) (*model.User, error) {
	if uid >= int64(len(s.storage)) {
		return nil, repository.ErrorUserNotFound
	}
	return &model.User{
		ID:   uid,
		Name: s.storage[uid],
	}, nil
}
