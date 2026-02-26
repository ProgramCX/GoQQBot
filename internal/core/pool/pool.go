// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package pool

import (
	"sync"

	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/internal/shared/types"
)

// Pool definition
type Pool struct {
	workers int
	queue   chan *types.Task
	wg      sync.WaitGroup
	handler func(task *types.Task)
}

// New Pool Constructor
func New(worker int, buffer int, handler func(task *types.Task)) *Pool {
	return &Pool{
		workers: worker,
		queue:   make(chan *types.Task, buffer),
		handler: handler,
	}
}

// Start the pool
func (p *Pool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// worker do tasks in queue
func (p *Pool) worker() {
	defer p.wg.Done()
	for task := range p.queue {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.ERROR("Task panic: %v", r)
				}
			}()
			p.handler(task)
		}()

		if task.Done != nil {
			task.Done()
		}
	}
}

func (p *Pool) Submit(task *types.Task) {
	select {
	case p.queue <- task:
	default:
		logger.WARN("Task queue full, dropping payload")
	}
}

func (p *Pool) Stop() {
	close(p.queue)
	p.wg.Wait()
}
