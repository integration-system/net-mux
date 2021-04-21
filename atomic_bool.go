package mux

import "sync/atomic"

type atomicBool struct {
	v *uint32
}

func (b *atomicBool) Get() bool {
	return atomic.LoadUint32(b.v) == 1
}

func (b *atomicBool) SetTrue() bool {
	return atomic.CompareAndSwapUint32(b.v, 0, 1)
}

func newAtomicBool() *atomicBool {
	b := &atomicBool{}
	v := uint32(0)
	b.v = &v
	return b
}
