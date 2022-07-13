/*
缓存值的抽象与封装
*/
package ccache

type ByteView struct {
	b []byte
}

func (bv ByteView) Len() int {
	return len(bv.b)
}

func (bv ByteView) String() string {
	return string(bv.b)
}

func (bv ByteView) ByteSlice() []byte {
	return cloneBytes(bv.b)
}

// 对缓存值进行拷贝，防止返回后外部对其有控制权
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
