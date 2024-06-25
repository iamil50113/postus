package user

import (
	"context"
	"fmt"
	"log/slog"
	"postus/internal/domain/model"
)

type User struct {
	log         *slog.Logger
	usrProvider UserProvider
}

type UserProvider interface {
	User(ctx context.Context, id int64) (*model.User, error)
}

func New(
	log *slog.Logger,
	usrProvider UserProvider) *User {
	return &User{
		log:         log,
		usrProvider: usrProvider,
	}
}

func (u *User) User(ctx context.Context, uid int64) (*model.User, error) {
	user, err := u.usrProvider.User(ctx, uid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("Invalid user id")
	}
	return user, err
}

func (u *User) UserIsExists(ctx context.Context, uid int64) bool {
	if user, err := u.usrProvider.User(ctx, uid); err != nil || user == nil {
		return false
	}
	return true
}
