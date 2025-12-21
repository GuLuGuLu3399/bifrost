package lifecycle

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Closer interface {
	Close() error
}

type Shutdown struct {
	mu      sync.Mutex
	closers []Closer
	once    sync.Once
}

func NewShutdown() *Shutdown {
	return &Shutdown{}
}

// Register 添加一个需要在关闭时释放的资源。
func (s *Shutdown) Register(c Closer) {
	if c == nil {
		return
	}
	s.mu.Lock()
	s.closers = append(s.closers, c)
	s.mu.Unlock()
}

// NotifyContext 返回一个在接收到中断信号时取消的 Context 以及取消函数。
func (s *Shutdown) NotifyContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(ch)
		close(ch)
	}()

	return ctx, cancel
}

// CloseAll 关闭所有注册的资源，按注册的相反顺序关闭，并返回所有关闭过程中遇到的错误。
func (s *Shutdown) CloseAll() error {
	var outErr error

	s.once.Do(func() {
		s.mu.Lock()
		closers := make([]Closer, len(s.closers))
		copy(closers, s.closers)
		s.mu.Unlock()

		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				outErr = errors.Join(outErr, err)
			}
		}
	})

	return outErr
}
