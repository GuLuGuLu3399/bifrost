package service

import (
	"context"
	"time"

	nexusv1 "github.com/gulugulu3399/bifrost/api/content/v1/nexus"
	"github.com/gulugulu3399/bifrost/internal/nexus/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

const tagWriteTimeout = 15 * time.Second

type TagService struct {
	nexusv1.UnimplementedTagServiceServer
	tagUc *biz.TagUseCase
}

func NewTagService(tagUc *biz.TagUseCase) *TagService {
	return &TagService{
		tagUc: tagUc,
	}
}

// DeleteTag 删除标签
func (s *TagService) DeleteTag(ctx context.Context, req *nexusv1.DeleteTagRequest) (*nexusv1.DeleteTagResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, tagWriteTimeout)
	defer cancel()

	if req.GetTagId() == 0 {
		return nil, xerr.New(xerr.CodeBadRequest, "tag_id is required")
	}

	err := s.tagUc.DeleteTag(ctx, req.GetTagId())
	if err != nil {
		return nil, err
	}

	return &nexusv1.DeleteTagResponse{Success: true}, nil
}
