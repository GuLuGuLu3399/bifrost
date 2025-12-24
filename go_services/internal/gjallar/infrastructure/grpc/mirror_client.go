package grpc

import (
	"context"
	"fmt"
	"time"

	commonv1 "github.com/gulugulu3399/bifrost/api/common/v1"
	searchv1 "github.com/gulugulu3399/bifrost/api/search/v1"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

// MirrorClient 封装 Mirror 搜索服务的 gRPC 客户端
type MirrorClient struct {
	client searchv1.MirrorServiceClient
}

// NewMirrorClient 创建 Mirror 客户端
func NewMirrorClient(client searchv1.MirrorServiceClient) *MirrorClient {
	return &MirrorClient{client: client}
}

// SearchRequest 搜索请求参数
type SearchRequest struct {
	Query      string
	Page       int32
	PageSize   int32
	CategoryID int64
	TagID      int64
	AuthorID   int64
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Hits       []*SearchHit
	TotalHits  int32
	TookMs     float32
	Facets     map[string]map[string]int64
}

// SearchHit 搜索结果项
type SearchHit struct {
	ID               int64
	Score            float32
	Title            string
	Slug             string
	HighlightTitle   string
	HighlightContent string
	PublishedAt      int64
}

// Search 执行全文搜索
// 将调用 Rust Mirror 服务获取搜索结果
func (c *MirrorClient) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 记录调用日志
	logger.WithContext(ctx).Info("Calling Rust Service: Mirror.Search",
		logger.String("query", req.Query),
		logger.Int32("page", req.Page),
		logger.Int32("page_size", req.PageSize),
	)

	// 构建 Filter
	filter := &searchv1.SearchRequest_Filter{}
	if req.CategoryID > 0 {
		filter.CategoryId = req.CategoryID
	}
	if req.TagID > 0 {
		filter.TagId = req.TagID
	}
	if req.AuthorID > 0 {
		filter.AuthorId = req.AuthorID
	}

	// 调用 Mirror gRPC 服务
	resp, err := c.client.Search(ctx, &searchv1.SearchRequest{
		Query: req.Query,
		Page: &commonv1.PageRequest{
			PageSize:  req.PageSize,
			PageToken: fmt.Sprintf("%d", req.Page), // 使用 page 作为 token
		},
		Filter:        filter,
		IncludeFacets: true, // 默认包含分面统计
	})

	if err != nil {
		logger.WithContext(ctx).Error("Mirror.Search failed",
			logger.Err(err),
			logger.String("query", req.Query),
		)
		return nil, xerr.Wrap(err, xerr.CodeInternal, "search service unavailable")
	}

	// 记录成功日志
	logger.WithContext(ctx).Info("Mirror.Search succeeded",
		logger.Int32("total_hits", resp.TotalHits),
		logger.Float32("took_ms", resp.TookMs),
	)

	// 转换结果
	hits := make([]*SearchHit, 0, len(resp.Hits))
	for _, h := range resp.Hits {
		hits = append(hits, &SearchHit{
			ID:               h.Id,
			Score:            h.Score,
			Title:            h.Title,
			Slug:             h.Slug,
			HighlightTitle:   h.HighlightTitle,
			HighlightContent: h.HighlightContent,
			PublishedAt:      h.PublishedAt,
		})
	}

	// 转换 Facets
	facets := make(map[string]map[string]int64)
	if resp.Facets != nil {
		if len(resp.Facets.Categories) > 0 {
			facets["categories"] = resp.Facets.Categories
		}
		if len(resp.Facets.Tags) > 0 {
			facets["tags"] = resp.Facets.Tags
		}
	}

	return &SearchResponse{
		Hits:      hits,
		TotalHits: resp.TotalHits,
		TookMs:    resp.TookMs,
		Facets:    facets,
	}, nil
}

// Suggest 获取搜索建议（自动补全）
func (c *MirrorClient) Suggest(ctx context.Context, prefix string, limit int32) ([]string, error) {
	if limit <= 0 {
		limit = 5
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	logger.WithContext(ctx).Info("Calling Rust Service: Mirror.Suggest",
		logger.String("prefix", prefix),
		logger.Int32("limit", limit),
	)

	resp, err := c.client.Suggest(ctx, &searchv1.SuggestRequest{
		Prefix: prefix,
		Limit:  limit,
	})

	if err != nil {
		logger.WithContext(ctx).Error("Mirror.Suggest failed",
			logger.Err(err),
			logger.String("prefix", prefix),
		)
		return nil, fmt.Errorf("suggest service unavailable: %w", err)
	}

	logger.WithContext(ctx).Info("Mirror.Suggest succeeded",
		logger.Int("suggestions_count", len(resp.Suggestions)),
		logger.Float32("took_ms", resp.TookMs),
	)

	return resp.Suggestions, nil
}
