package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	grpcClient "github.com/gulugulu3399/bifrost/internal/gjallar/infrastructure/grpc"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

// SearchHandler 处理搜索请求
type SearchHandler struct {
	mirrorClient *grpcClient.MirrorClient
}

// NewSearchHandler 创建搜索处理器
func NewSearchHandler(mirrorClient *grpcClient.MirrorClient) *SearchHandler {
	return &SearchHandler{mirrorClient: mirrorClient}
}

// SearchResponse API 响应格式
type SearchResponse struct {
	Hits      []SearchHitDTO         `json:"hits"`
	TotalHits int32                  `json:"total_hits"`
	TookMs    float32                `json:"took_ms"`
	Facets    map[string]map[string]int64 `json:"facets,omitempty"`
}

// SearchHitDTO 搜索结果 DTO
type SearchHitDTO struct {
	ID               int64   `json:"id"`
	Score            float32 `json:"score"`
	Title            string  `json:"title"`
	Slug             string  `json:"slug"`
	HighlightTitle   string  `json:"highlight_title,omitempty"`
	HighlightContent string  `json:"highlight_content,omitempty"`
	PublishedAt      int64   `json:"published_at"`
}

// ServeHTTP 实现 http.Handler 接口
func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if h.mirrorClient == nil {
		http.Error(w, `{"error":"search service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	// 解析查询参数
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		query = strings.TrimSpace(r.URL.Query().Get("query"))
	}

	page, ok := parseOptionalInt32(r.URL.Query().Get("page"), 1)
	if !ok {
		writeBadRequest(w, "invalid page")
		return
	}
	if page < 1 {
		writeBadRequest(w, "page must be >= 1")
		return
	}

	pageSize, ok := parseOptionalInt32(r.URL.Query().Get("page_size"), 20)
	if !ok {
		writeBadRequest(w, "invalid page_size")
		return
	}
	if pageSize < 1 || pageSize > 100 {
		writeBadRequest(w, "page_size must be between 1 and 100")
		return
	}

	categoryID, ok := parseOptionalInt64(r.URL.Query().Get("category_id"), 0)
	if !ok {
		writeBadRequest(w, "invalid category_id")
		return
	}
	tagID, ok := parseOptionalInt64(r.URL.Query().Get("tag_id"), 0)
	if !ok {
		writeBadRequest(w, "invalid tag_id")
		return
	}
	authorID, ok := parseOptionalInt64(r.URL.Query().Get("author_id"), 0)
	if !ok {
		writeBadRequest(w, "invalid author_id")
		return
	}

	// 记录请求日志
	logger.WithContext(ctx).Info("Search request received",
		logger.String("query", query),
		logger.Int32("page", page),
		logger.Int32("page_size", pageSize),
	)

	// 调用 Mirror 服务
	resp, err := h.mirrorClient.Search(ctx, &grpcClient.SearchRequest{
		Query:      query,
		Page:       page,
		PageSize:   pageSize,
		CategoryID: categoryID,
		TagID:      tagID,
		AuthorID:   authorID,
	})

	if err != nil {
		logger.WithContext(ctx).Error("Search failed", logger.Err(err))
		http.Error(w, `{"error":"search service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	// 转换为 DTO
	hits := make([]SearchHitDTO, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		hits = append(hits, SearchHitDTO{
			ID:               hit.ID,
			Score:            hit.Score,
			Title:            hit.Title,
			Slug:             hit.Slug,
			HighlightTitle:   hit.HighlightTitle,
			HighlightContent: hit.HighlightContent,
			PublishedAt:      hit.PublishedAt,
		})
	}

	// 构建响应
	result := SearchResponse{
		Hits:      hits,
		TotalHits: resp.TotalHits,
		TookMs:    resp.TookMs,
		Facets:    resp.Facets,
	}

	// 返回 JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.WithContext(ctx).Error("Failed to encode response", logger.Err(err))
	}
}

// SuggestHandler 处理搜索建议请求
type SuggestHandler struct {
	mirrorClient *grpcClient.MirrorClient
}

// NewSuggestHandler 创建搜索建议处理器
func NewSuggestHandler(mirrorClient *grpcClient.MirrorClient) *SuggestHandler {
	return &SuggestHandler{mirrorClient: mirrorClient}
}

// SuggestResponse API 响应格式
type SuggestResponse struct {
	Suggestions []string `json:"suggestions"`
	TookMs      float32  `json:"took_ms,omitempty"`
}

// ServeHTTP 实现 http.Handler 接口
func (h *SuggestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if h.mirrorClient == nil {
		http.Error(w, `{"error":"suggest service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	// 解析查询参数
	prefix := strings.TrimSpace(r.URL.Query().Get("prefix"))
	if prefix == "" {
		prefix = strings.TrimSpace(r.URL.Query().Get("q"))
	}

	limit, ok := parseOptionalInt32(r.URL.Query().Get("limit"), 5)
	if !ok {
		writeBadRequest(w, "invalid limit")
		return
	}
	if limit < 1 || limit > 20 {
		writeBadRequest(w, "limit must be between 1 and 20")
		return
	}

	// 调用 Mirror 服务
	suggestions, err := h.mirrorClient.Suggest(ctx, prefix, limit)
	if err != nil {
		logger.WithContext(ctx).Error("Suggest failed", logger.Err(err))
		http.Error(w, `{"error":"suggest service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	// 构建响应
	result := SuggestResponse{
		Suggestions: suggestions,
	}

	// 返回 JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.WithContext(ctx).Error("Failed to encode response", logger.Err(err))
	}
}

// 辅助函数

func parseOptionalInt32(s string, defaultValue int32) (int32, bool) {
	n := strings.TrimSpace(s)
	if n == "" {
		return defaultValue, true
	}
	v, err := strconv.ParseInt(n, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(v), true
}

func parseOptionalInt64(s string, defaultValue int64) (int64, bool) {
	n := strings.TrimSpace(s)
	if n == "" {
		return defaultValue, true
	}
	v, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func writeBadRequest(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
