// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

package pool

import (
	"context"
	"errors"
	"runtime/debug"
	"sync"

	"github.com/ProgramCX/GoQQBot/shared/types"
)

// Submission errors distinguish overload from a closed pool.
// 提交错误区分过载和线程池已关闭。
var (
	ErrFull   = errors.New("event queue is full")
	ErrClosed = errors.New("event pool is closed")
)

// Pool processes bounded concurrent tasks. Start and Shutdown are idempotent.
// Pool 有界并发处理任务；Start 和 Shutdown 均幂等。
type Pool struct {
	workers int
	queue   chan *types.Task
	handler func(*types.Task)
	mu      sync.Mutex
	started bool
	closed  bool
	done    chan struct{}
	onPanic func(*types.Task, *types.PanicError)
}

// New creates a dormant pool. The caller must supply positive sizes and a handler.
// New 创建尚未启动的线程池，调用方必须提供正数容量和处理函数。
// The optional panic observer must not panic; Bot supplies a protected observer retaining failures.
// 可选 panic 观察函数不得 panic；Bot 会提供保留故障的受保护观察函数。
func New(workers, capacity int, handler func(*types.Task), observers ...func(*types.Task, *types.PanicError)) *Pool {
	p := &Pool{workers: workers, queue: make(chan *types.Task, capacity), handler: handler, done: make(chan struct{})}
	if len(observers) > 0 {
		p.onPanic = observers[0]
	}
	return p
}

// Start starts workers once. Calling Start after Shutdown does nothing.
// Start 仅启动一次工作协程；Shutdown 后调用 Start 不执行操作。
func (p *Pool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.started || p.closed {
		return
	}
	p.started = true
	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); p.worker() }()
	}
	go func() { wg.Wait(); close(p.done) }()
}

func (p *Pool) worker() {
	for task := range p.queue {
		// Panics are contained; completion allows progress even for a failed handler.
		// panic 会被隔离；即使处理函数失败，完成通知仍可推进进度。
		func() {
			defer func() {
				if value := recover(); value != nil && p.onPanic != nil {
					p.onPanic(task, &types.PanicError{Value: value, Stack: debug.Stack()})
				}
			}()
			defer func() {
				if task.Done != nil {
					task.Done()
				}
			}()
			p.handler(task)
		}()
	}
}

// Submit accepts a task without blocking. No task is accepted after Shutdown begins.
// Submit 非阻塞接收任务；Shutdown 开始后不再接收新任务。
func (p *Pool) Submit(task *types.Task) error {
	if task == nil {
		return errors.New("nil event task")
	}
	if task.Ctx != nil {
		if err := task.Ctx.Err(); err != nil {
			return err
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || !p.started {
		return ErrClosed
	}
	select {
	case p.queue <- task:
		return nil
	default:
		return ErrFull
	}
}

// Shutdown rejects new tasks and drains accepted work until ctx expires.
// Shutdown 拒绝新任务，并在 ctx 到期前处理已接受的任务。
// A callback that ignores cancellation may continue after Shutdown returns.
// 忽略取消信号的回调可能在 Shutdown 返回后继续运行。
func (p *Pool) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	if !p.closed {
		p.closed = true
		close(p.queue)
		if !p.started {
			close(p.done)
		}
	}
	p.mu.Unlock()
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop waits indefinitely for accepted work. Prefer Shutdown with a deadline.
// Stop 无限等待已接受任务完成；建议优先使用带截止时间的 Shutdown。
func (p *Pool) Stop() { _ = p.Shutdown(context.Background()) }
