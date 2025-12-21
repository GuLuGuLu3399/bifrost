package cache

import "testing"

func TestClient_Close_NilSafe(t *testing.T) {
	var c *Client
	if err := c.Close(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	c = &Client{}
	if err := c.Close(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
