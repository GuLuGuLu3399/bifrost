package data

import (
	"fmt"
	"time"
)

const (
	PostCacheTTL = 1 * time.Hour
	UserCacheTTL = 2 * time.Hour
	MetaCacheTTL = 24 * time.Hour // 分类/标签变动极少
)

func KeyPostDetail(slugOrID string) string { return fmt.Sprintf("beacon:post:%s", slugOrID) }
func KeyUserProfile(userID int64) string   { return fmt.Sprintf("beacon:user:%d", userID) }
func KeyCategoryList() string              { return "beacon:categories:all" }
func KeyTagList() string                   { return "beacon:tags:popular" }
