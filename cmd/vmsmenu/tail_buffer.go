package main

// newTailBuffer creates a new tailBuffer that keeps up to max bytes.
//
// Used to capture the tail of command output.
func newTailBuffer(max int) *tailBuffer {
	if max <= 0 {
		max = 4096
	}
	return &tailBuffer{max: max}
}

// Write implements io.Writer for tailBuffer.
func (t *tailBuffer) Write(p []byte) (int, error) {
	// nothing to do
	if len(p) == 0 {
		return 0, nil
	}

	// too much data, just keep the tail
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

// String returns the contents of the buffer as a Go string.
func (t *tailBuffer) String() string {
	return string(t.buf)
}
