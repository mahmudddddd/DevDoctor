package runner

import (
	"bytes"
	"io"
	"strings"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

const captureBufferSize = 32 * 1024

type captureOutcome struct {
	capture model.StreamCapture
	err     error
}

type boundedCapture struct {
	retained bytes.Buffer
	total    int64
	limit    int64
}

func newBoundedCapture(limit int64) *boundedCapture {
	return &boundedCapture{limit: limit}
}

func (capture *boundedCapture) Write(data []byte) (int, error) {
	capture.total += int64(len(data))
	remaining := capture.limit - int64(capture.retained.Len())
	if remaining > 0 {
		keep := int64(len(data))
		if keep > remaining {
			keep = remaining
		}
		_, _ = capture.retained.Write(data[:keep])
	}
	return len(data), nil
}

func (capture *boundedCapture) Result() model.StreamCapture {
	return finalizeCapture(capture.retained.Bytes(), capture.total, capture.limit)
}

func captureStream(reader io.Reader, limit int64) captureOutcome {
	var retained bytes.Buffer
	buffer := make([]byte, captureBufferSize)
	var total int64

	for {
		count, err := reader.Read(buffer)
		if count > 0 {
			total += int64(count)
			remaining := limit - int64(retained.Len())
			if remaining > 0 {
				keep := int64(count)
				if keep > remaining {
					keep = remaining
				}
				_, _ = retained.Write(buffer[:keep])
			}
		}
		if err != nil {
			capture := finalizeCapture(retained.Bytes(), total, limit)
			if err == io.EOF {
				return captureOutcome{capture: capture}
			}
			return captureOutcome{capture: capture, err: err}
		}
	}
}

func finalizeCapture(retained []byte, total, limit int64) model.StreamCapture {
	captured := int64(len(retained))
	discarded := total - captured
	if discarded < 0 {
		discarded = 0
	}
	return model.StreamCapture{
		Text:           strings.ToValidUTF8(string(retained), "�"),
		CapturedBytes:  captured,
		DiscardedBytes: discarded,
		Truncated:      total > limit,
	}
}
