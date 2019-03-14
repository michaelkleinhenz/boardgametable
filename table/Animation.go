package table

// Animation represents an animation.
type Animation interface {
	Step()
	SetFrameBuffer(frameBuffer *[]byte)
	GetFrameBuffer() *[]byte
}