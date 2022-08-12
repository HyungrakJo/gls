package gls

// though this could probably be better at keeping ids smaller, the goal of
// this class is to keep a registry of the smallest unique integer ids
// per-process possible

import (
	"sync/atomic"

	"golang.design/x/lockfree"
)

type idPool struct {
	queue *lockfree.Queue
	curID uint32
}

func (p *idPool) newID() uint32 {
	curID := atomic.AddUint32(&p.curID, 1)
	return curID - 1
}

func (p *idPool) Acquire() (id uint32) {
	if item := p.queue.Dequeue(); item != nil {
		return item.(uint32)
	} else {
		return p.newID()
	}
}

func (p *idPool) Release(id uint32) {
	p.queue.Enqueue(id)
}
