package main

import (
	"context"
	"os"

	"github.com/gulugulu3399/bifrost/internal/pkg/lifecycle"
)

func main() {
	sh := lifecycle.NewShutdown()
	ctx, stop := sh.NotifyContext(context.Background())
	defer stop()

	// TODO: init resources and servers
	<-ctx.Done()
	_ = sh.CloseAll()
	os.Exit(0)
}
