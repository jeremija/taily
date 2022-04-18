package action

import (
	"bytes"
	"sync"
)

type pool struct {
	pool sync.Pool
}

func newPool() *pool {
	return &pool{
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func (p *pool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

func (p *pool) Put(b *bytes.Buffer) {
	b.Reset()
	p.pool.Put(b)
}
