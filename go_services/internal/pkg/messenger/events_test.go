package messenger

import (
	"encoding/json"
	"testing"
)

func TestPostEventPayload_JSONContract_Full(t *testing.T) {
	p := PostEventPayload{
		ID:          123,
		Slug:        "hello-bifrost",
		Title:       "Hello Bifrost",
		Summary:     "summary",
		Status:      2,
		PublishedAt: 1760000000,
	}

	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal map failed: %v", err)
	}

	required := []string{"id", "slug", "title", "summary", "status", "published_at"}
	for _, key := range required {
		if _, ok := got[key]; !ok {
			t.Fatalf("expected key %q in payload: %s", key, string(b))
		}
	}
}

func TestPostEventPayload_JSONContract_MinimalCompat(t *testing.T) {
	var p PostEventPayload
	if err := json.Unmarshal([]byte(`{"id":123}`), &p); err != nil {
		t.Fatalf("unmarshal minimal payload failed: %v", err)
	}
	if p.ID != 123 {
		t.Fatalf("unexpected id: %d", p.ID)
	}
	if p.Title != "" || p.Slug != "" {
		t.Fatalf("expected optional fields to be zero-value, got title=%q slug=%q", p.Title, p.Slug)
	}
}

func TestSubjectConstants(t *testing.T) {
	if SubjectPostCreated != "content.post.created" {
		t.Fatalf("SubjectPostCreated mismatch: %s", SubjectPostCreated)
	}
	if SubjectPostUpdated != "content.post.updated" {
		t.Fatalf("SubjectPostUpdated mismatch: %s", SubjectPostUpdated)
	}
	if SubjectPostDeleted != "content.post.deleted" {
		t.Fatalf("SubjectPostDeleted mismatch: %s", SubjectPostDeleted)
	}
	if SubjectPostWildcard != "content.post.>" {
		t.Fatalf("SubjectPostWildcard mismatch: %s", SubjectPostWildcard)
	}
	if StreamContent != "BIFROST_CONTENT" {
		t.Fatalf("StreamContent mismatch: %s", StreamContent)
	}
	if ConsumerMirrorIndexer != "mirror_indexer" {
		t.Fatalf("ConsumerMirrorIndexer mismatch: %s", ConsumerMirrorIndexer)
	}
}
