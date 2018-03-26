package bytes

import (
	"sync"
)

//pool内存组织如下，pool是一个链式存储的栈，数据从栈顶出，同时数据也从栈顶回收。栈的链式存储结构，简称链栈。

type Buffer struct {
	buf  []byte
	next *Buffer // next free buffer
}

func (b *Buffer) Bytes() []byte {
	return b.buf
}

// Pool is a buffer pool.
type Pool struct {
	lock sync.Mutex
	free *Buffer
	max  int
	num  int
	size int
}

// NewPool new a memory buffer pool struct.
func NewPool(num, size int) (p *Pool) {
	p = new(Pool)
	p.init(num, size)
	return
}

// Init init the memory buffer.
func (p *Pool) Init(num, size int) {
	p.init(num, size)
	return
}

// init init the memory buffer.
func (p *Pool) init(num, size int) {
	p.num = num
	p.size = size
	p.max = num * size
	p.grow()
}

// grow grow the memory buffer size, and update free pointer.重置Pool对象
func (p *Pool) grow() {
	var (
		i   int
		b   *Buffer
		bs  []Buffer
		buf []byte
	)
	buf = make([]byte, p.max)
	bs = make([]Buffer, p.num)
	p.free = &bs[0]
	b = p.free
	for i = 1; i < p.num; i++ {
		b.buf = buf[(i-1)*p.size : i*p.size]
		b.next = &bs[i]
		b = b.next
	}
	b.buf = buf[(i-1)*p.size : i*p.size]
	b.next = nil
	return
}

// Get get a free memory buffer.获取一个缓冲区
func (p *Pool) Get() (b *Buffer) {
	p.lock.Lock()
	if b = p.free; b == nil {
		p.grow()
		b = p.free
	}
	p.free = b.next
	p.lock.Unlock()
	return
}

// Put put back a memory buffer to free.归还一个缓冲区
func (p *Pool) Put(b *Buffer) {
	p.lock.Lock()
	b.next = p.free
	p.free = b
	p.lock.Unlock()
	return
}
