package output

import (
	"bytes"
	"errors"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer

	if err := WriteJSON(&buf, map[string]any{"ok": true}); err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	want := "{\n  \"ok\": true\n}\n"
	if buf.String() != want {
		t.Fatalf("unexpected JSON output:\n%s", buf.String())
	}
}

func TestWriteErrorJSON(t *testing.T) {
	var buf bytes.Buffer

	if err := WriteError(&buf, true, errors.New("boom")); err != nil {
		t.Fatalf("WriteError returned error: %v", err)
	}

	want := "{\n  \"error\": \"boom\"\n}\n"
	if buf.String() != want {
		t.Fatalf("unexpected error output:\n%s", buf.String())
	}
}
