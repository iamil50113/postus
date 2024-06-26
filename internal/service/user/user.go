package user

import (
	"context"
	"errors"
	"log/slog"
	"postus/internal/domain/model"
	"postus/internal/repository"
	"postus/internal/service"
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
		if errors.Is(err, repository.ErrorUserNotFound) {
			return nil, err
		} else {
			return nil, service.ErrorServer
		}
	}
	return user, err
}

//func (u *User) UserIsExists(ctx context.Context, uid int64) (bool, error) {
//	if _, err := u.User(ctx, uid); err != nil {
//		if err != nil {
//			if errors.Is(err, repository.ErrorUserNotFound) {
//				return false, nil
//			} else {
//				return false, err
//			}
//		}
//	}
//
//	return true, nil
//}
