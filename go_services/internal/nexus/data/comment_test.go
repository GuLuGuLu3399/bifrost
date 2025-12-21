package data

import (
	"testing"
	"time"

	contentv1 "github.com/gulugulu3399/bifrost/api/content/v1"
)

func TestCommentPO_ToEntity_UnknownStatusFallback(t *testing.T) {
	po := &commentPO{
		ID:        123,
		PostID:    1,
		UserID:    2,
		Content:   "hi",
		Status:    "weird_status",
		Version:   7,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	c := po.toEntity()
	if c == nil {
		t.Fatalf("expected entity")
	}
	if c.Status != contentv1.CommentStatus_COMMENT_STATUS_PENDING {
		t.Fatalf("expected fallback status pending, got %v", c.Status)
	}
	if c.Version != 7 {
		t.Fatalf("expected version preserved")
	}
}
