package connect

// TailBuffer captures the last N bytes written to it.
//
// It implements io.Writer.
type TailBuffer struct {
	buf []byte // buffered data
	max int    // maximum bytes to keep
}

// String returns the buffer contents as a string.
func (t *TailBuffer) String() string {
	return string(t.buf)
}

// NewTailBuffer creates a new TailBuffer that keeps up to max bytes.
func NewTailBuffer(max int) *TailBuffer {
	if max <= 0 {
		max = 4096
	}
	return &TailBuffer{max: max}
}

// Write implements io.Writer.
func (t *TailBuffer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	// too much data, keep only the tail
	if len(p) >= t.max {
		t.buf = append(t.buf[:0], p[len(p)-t.max:]...)
		return len(p), nil
	}

	// append and trim from the front if needed
	need := len(t.buf) + len(p) - t.max
	if need > 0 {
		t.buf = t.buf[need:]
	}
	t.buf = append(t.buf, p...)
	return len(p), nil
}
