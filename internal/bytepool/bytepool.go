package bytepool

import "sync"

type Buffer struct {
	Data []byte
	pool *sync.Pool
}

func (b *Buffer) Pooled() bool {
	return b.pool != nil
}

func (b *Buffer) Free() {
	if b.pool != nil {
		b.pool.Put(b.Data)
	}

	b.Data = nil
	b.pool = nil
}

type Pool struct {
	pool    *sync.Pool
	maxSize int
}

func New(maxSize int) *Pool {
	return &Pool{
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, maxSize)
			},
		},
		maxSize: maxSize,
	}
}

func (p *Pool) Get(size int) Buffer {
	if size <= p.maxSize {
		return Buffer{
			Data: p.pool.Get().([]byte)[:size],
			pool: p.pool,
		}
	}

	return Buffer{
		Data: make([]byte, size),
	}
}
