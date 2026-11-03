package pool

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ProgramCX/GoQQBot/shared/types"
)

func TestOverloadDrainAndIdempotentShutdown(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	p := New(1, 1, func(*types.Task) { once.Do(func() { close(started) }); <-release })
	p.Start()
	p.Start()
	if err := p.Submit(&types.Task{}); err != nil {
		t.Fatal(err)
	}
	<-started
	if err := p.Submit(&types.Task{}); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(&types.Task{}); !errors.Is(err, ErrFull) {
		t.Fatalf("want overload, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.Shutdown(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := p.Submit(&types.Task{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("want closed, got %v", err)
	}
	close(release)
	wait, finish := context.WithTimeout(context.Background(), time.Second)
	defer finish()
	if err := p.Shutdown(wait); err != nil {
		t.Fatal(err)
	}
	if err := p.Shutdown(wait); err != nil {
		t.Fatal(err)
	}
}
func TestPanicDoesNotKillWorker(t *testing.T) {
	done := make(chan struct{}, 2)
	p := New(1, 2, func(*types.Task) { panic("handler failed") })
	p.Start()
	for range 2 {
		if err := p.Submit(&types.Task{Done: func() { done <- struct{}{} }}); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if len(done) != 2 {
		t.Fatal("panic lost completion")
	}
}
