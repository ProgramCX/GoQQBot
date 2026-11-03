package adapter

import "sync"

type session struct {
	id        string
	watermark *dispatchWatermark
	authRetry bool
}

// invalidate drops resumable session state, without granting another token refresh.
// invalidate 丢弃可恢复会话状态，但不会重新允许令牌刷新重试。
func (s *session) invalidate() {
	s.id = ""
	s.watermark = &dispatchWatermark{}
}

// beginConnection keeps only completed delivery progress. Old callbacks retain the
// beginConnection 仅保留已完成的投递进度；旧回调保留
// old watermark, so late completion after disconnect cannot skip replayed events.
// 旧水位线，因此断开连接后的迟到完成通知不会跳过重放事件。
func (s *session) beginConnection() *int64 {
	var checkpoint *int64
	if s.watermark != nil {
		checkpoint = s.watermark.checkpoint()
	}
	s.watermark = &dispatchWatermark{seq: checkpoint}
	return checkpoint
}

type dispatchCompletion struct {
	seq  *int64
	done bool
}

// dispatchWatermark bounds outstanding deliveries and advances the resume sequence
// dispatchWatermark 限制待完成投递数量，并仅在按投递顺序
// only across completed entries in delivery order (sequence numbers may have gaps).
// 连续完成的条目间推进恢复序号；序号本身可以不连续。
// Registration is serialized by the connection loop; callbacks may finish concurrently.
// 连接循环串行登记投递，完成回调则可并发执行。
// Capacity checking and registration share a lock so the limit cannot be exceeded.
// 容量检查与登记共用一把锁，避免超过窗口上限。
type dispatchWatermark struct {
	mu      sync.Mutex
	seq     *int64
	pending []*dispatchCompletion
}

func (p *dispatchWatermark) checkpoint() *int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.seq == nil {
		return nil
	}
	value := *p.seq
	return &value
}

func (p *dispatchWatermark) tryAdd(seq *int64, limit int) (func(), bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.pending) >= limit {
		return nil, false
	}
	item := &dispatchCompletion{}
	if seq != nil {
		value := *seq
		item.seq = &value
	}
	p.pending = append(p.pending, item)
	return func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		item.done = true
		for len(p.pending) > 0 && p.pending[0].done {
			if p.pending[0].seq != nil {
				p.seq = p.pending[0].seq
			}
			p.pending[0] = nil
			p.pending = p.pending[1:]
		}
	}, true
}
