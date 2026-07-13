package runner

import (
	"errors"
	"strings"
	"testing"
)

func TestCaptureStreamRetainsPrefixAndCountsDiscardedBytes(t *testing.T) {
	t.Parallel()
	outcome := captureStream(strings.NewReader("abcdefghij"), 4)
	if outcome.err != nil {
		t.Fatal(outcome.err)
	}
	if outcome.capture.Text != "abcd" || outcome.capture.CapturedBytes != 4 || outcome.capture.DiscardedBytes != 6 || !outcome.capture.Truncated {
		t.Fatalf("capture = %+v", outcome.capture)
	}
}

func TestCaptureStreamConvertsPartialUTF8Deterministically(t *testing.T) {
	t.Parallel()
	outcome := captureStream(strings.NewReader("A€B"), 2)
	if outcome.err != nil {
		t.Fatal(outcome.err)
	}
	if outcome.capture.Text != "A�" || outcome.capture.CapturedBytes != 2 || outcome.capture.DiscardedBytes != 3 {
		t.Fatalf("capture = %+v", outcome.capture)
	}
}

func TestCaptureStreamPropagatesReadError(t *testing.T) {
	t.Parallel()
	want := errors.New("read failed")
	outcome := captureStream(errorReader{err: want}, 32)
	if !errors.Is(outcome.err, want) {
		t.Fatalf("error = %v, want %v", outcome.err, want)
	}
}

func TestBoundedCaptureNeverRetainsPastLimit(t *testing.T) {
	t.Parallel()
	capture := newBoundedCapture(1024)
	data := strings.Repeat("x", 1024*1024)
	if count, err := capture.Write([]byte(data)); err != nil || count != len(data) {
		t.Fatalf("Write() = %d, %v", count, err)
	}
	result := capture.Result()
	if result.CapturedBytes != 1024 || result.DiscardedBytes != int64(len(data)-1024) || !result.Truncated {
		t.Fatalf("result = %+v", result)
	}
}

type errorReader struct {
	err error
}

func (reader errorReader) Read([]byte) (int, error) {
	return 0, reader.err
}
