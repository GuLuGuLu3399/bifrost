package service

import (
	"context"

	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	"github.com/gulugulu3399/bifrost/internal/beacon/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

type UserService struct {
	repo biz.UserRepo
}

func NewUserService(repo biz.UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(ctx context.Context, req *beaconv1.GetUserRequest) (*beaconv1.GetUserResponse, error) {
	if req.GetUserId() == 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "user_id is required")
	}

	user, err := s.repo.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	return &beaconv1.GetUserResponse{User: user}, nil
}
