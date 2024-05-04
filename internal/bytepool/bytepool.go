package bytepool

import "sync"

// Buffer is a reusable byte slice. Data must not be modified after creation.
// The buffer must be released after use by calling Free.
type Buffer struct {
	Data []byte
	pool *sync.Pool
}

// Pooled returns true if the buffer was created from a pool.
func (b *Buffer) Pooled() bool {
	return b.pool != nil
}

// Free returns the buffer to the pool if it was created from one.
// The buffer must not be used after calling Free.
func (b *Buffer) Free() {
	if b.pool != nil {
		b.pool.Put(b.Data)
	}

	b.Data = nil
	b.pool = nil
}

// BytePool manages a pool of short reusable byte slices of up to maxSize bytes
// and allows for occasional allocation of larger slices when needed.
type BytePool struct {
	pool    *sync.Pool
	maxSize int
}

// New creates a new BytePool with maxSize slice size.
func New(maxSize int) *BytePool {
	return &BytePool{
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, maxSize)
			},
		},
		maxSize: maxSize,
	}
}

// Buffer returns a new Buffer of size bytes. If the requested size is larger
// than the allowed maximum, a new slice is allocated. The returned buffer
// contains garbage data and must be filled before use.
func (p *BytePool) Buffer(size int) Buffer {
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
