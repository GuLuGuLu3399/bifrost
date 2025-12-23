package service

import (
	"strconv"

	commonv1 "github.com/gulugulu3399/bifrost/api/common/v1"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
)

// parsePage 解析 commonv1.PageRequest。
// 约定：PageToken 作为页码字符串（"1", "2"...），解析失败则回退默认值。
func parsePage(p *commonv1.PageRequest) (page int, pageSize int) {
	page = defaultPage
	pageSize = defaultPageSize
	if p == nil {
		return
	}

	if p.GetPageSize() > 0 {
		pageSize = int(p.GetPageSize())
	}
	if tok := p.GetPageToken(); tok != "" {
		if n, err := strconv.Atoi(tok); err == nil && n > 0 {
			page = n
		}
	}
	return
}

// nextPageTokenByTotal 简单页码策略：如果还有更多数据，返回下一页 token，否则返回空字符串。
func nextPageTokenByTotal(page, pageSize int, total int64) string {
	if page < 1 {
		page = defaultPage
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if int64(page*pageSize) < total {
		return strconv.Itoa(page + 1)
	}
	return ""
}

// totalToInt32 将总数安全转换为 int32（溢出时 clamp 到 MaxInt32）。
func totalToInt32(total int64) int32 {
	const maxInt32 = int64(^uint32(0) >> 1)
	if total <= 0 {
		return 0
	}
	if total > maxInt32 {
		return int32(maxInt32)
	}
	return int32(total)
}
