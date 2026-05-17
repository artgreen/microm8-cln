package filerecord

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// --- Constructor -----------------------------------------------------------

func TestNewFileRecord_InitializesEmpty(t *testing.T) {
	t.Parallel()
	fr := NewFileRecord("/some/path", "file.txt")
	if fr.FilePath != "/some/path" {
		t.Errorf("FilePath = %q, want %q", fr.FilePath, "/some/path")
	}
	if fr.FileName != "file.txt" {
		t.Errorf("FileName = %q, want %q", fr.FileName, "file.txt")
	}
	if len(fr.Content) != 0 {
		t.Errorf("Content len = %d, want 0", len(fr.Content))
	}
	// Empty slice, not nil — the constructor explicitly allocates an empty slice.
	if fr.Content == nil {
		t.Error("Content should be allocated empty slice, not nil")
	}
}

// --- AddMeta / GetMeta -----------------------------------------------------

func TestAddMeta_LazyInitMap(t *testing.T) {
	t.Parallel()
	fr := NewFileRecord("/", "f")
	if fr.MetaData != nil {
		t.Fatal("test precondition: MetaData should start nil")
	}
	fr.AddMeta("key", "value")
	if got := fr.MetaData["key"]; got != "value" {
		t.Errorf("MetaData[key] = %q, want %q", got, "value")
	}
}

func TestAddMeta_EmptyValueDeletes(t *testing.T) {
	t.Parallel()
	fr := NewFileRecord("/", "f")
	fr.AddMeta("k", "v")
	fr.AddMeta("k", "")
	if _, ok := fr.MetaData["k"]; ok {
		t.Errorf("MetaData[k] still present after empty-value Add, want deleted")
	}
}

func TestGetMeta_ReturnsDefaultForMissing(t *testing.T) {
	t.Parallel()
	fr := NewFileRecord("/", "f")
	if got := fr.GetMeta("missing", "default"); got != "default" {
		t.Errorf("GetMeta(missing, default) = %q, want %q", got, "default")
	}
}

func TestGetMeta_ReturnsValueWhenPresent(t *testing.T) {
	t.Parallel()
	fr := NewFileRecord("/", "f")
	fr.AddMeta("k", "v")
	if got := fr.GetMeta("k", "default"); got != "v" {
		t.Errorf("GetMeta(k, default) = %q, want %q", got, "v")
	}
}

// --- Content read/write ----------------------------------------------------

func TestWrite_AppendsByte(t *testing.T) {
	t.Parallel()
	fr := NewFileRecord("/", "f")
	for _, b := range []byte("hello") {
		fr.Write(b)
	}
	if !bytes.Equal(fr.Content, []byte("hello")) {
		t.Errorf("Content = %q, want %q", fr.Content, "hello")
	}
}

func TestWriteBytes_AppendsAll(t *testing.T) {
	t.Parallel()
	fr := NewFileRecord("/", "f")
	fr.WriteBytes([]byte("hello "))
	fr.WriteBytes([]byte("world"))
	if !bytes.Equal(fr.Content, []byte("hello world")) {
		t.Errorf("Content = %q, want %q", fr.Content, "hello world")
	}
}

func TestReadBytes_DrainsInChunks(t *testing.T) {
	t.Parallel()
	fr := NewFileRecord("/", "f")
	fr.WriteBytes([]byte("hello world"))

	buf := make([]byte, 5)
	n, err := fr.ReadBytes(buf)
	if err != nil {
		t.Fatalf("first read: %v", err)
	}
	if n != 5 || !bytes.Equal(buf, []byte("hello")) {
		t.Errorf("first read: got n=%d buf=%q, want n=5 buf=%q", n, buf, "hello")
	}

	n, err = fr.ReadBytes(buf)
	if err != nil {
		t.Fatalf("second read: %v", err)
	}
	if n != 5 || !bytes.Equal(buf, []byte(" worl")) {
		t.Errorf("second read: got n=%d buf=%q, want n=5 buf=%q", n, buf, " worl")
	}
}

// --- JSON roundtrip: exercises the Phase-0 struct-tag fix ------------------
//
// Before Phase 0:  ID bson.ObjectId `bson:"_id,omitempty";json:"-"`  (broken)
// After:           ID bson.ObjectId `bson:"_id,omitempty" json:"-"`  (correct)
//
// The malformed tag made reflect.StructTag.Get silently return empty for
// both bson and json keys on the ID field. We assert that:
//   1. JSON marshaling no longer includes any "ID" field.
//   2. The other tagged fields (filename, path, content, etc.) marshal
//      with their expected names.

func TestFileRecord_JSON_DoesNotIncludeID(t *testing.T) {
	t.Parallel()
	fr := NewFileRecord("/some/path", "file.txt")
	fr.Description = "test desc"
	fr.Content = []byte("hello")
	fr.ContentSize = 5

	data := fr.JSON()
	if len(data) == 0 {
		t.Fatal("JSON() returned empty bytes")
	}

	var generic map[string]interface{}
	if err := json.Unmarshal(data, &generic); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, data)
	}

	// ID has json:"-" so it must NOT appear in output.
	for k := range generic {
		if strings.EqualFold(k, "ID") || strings.EqualFold(k, "_id") {
			t.Errorf("JSON output contains forbidden key %q (struct tag must hide ID)", k)
		}
	}

	// Tagged fields must appear with their lowercase names.
	mustHave := []string{"filename", "path", "content", "size", "description", "dir", "deleted"}
	for _, k := range mustHave {
		if _, ok := generic[k]; !ok {
			t.Errorf("JSON output missing required key %q", k)
		}
	}
}

func TestFileRecord_JSON_Roundtrip(t *testing.T) {
	t.Parallel()
	src := NewFileRecord("/some/path", "file.txt")
	src.Description = "round-trip test"
	src.ContentSize = 5
	src.Content = []byte("hello")
	src.Address = 0x1000
	src.Directory = true

	data := src.JSON()

	dst := &FileRecord{}
	dst.UnJSON(data)

	// Compare only the fields that marshal — those are the ones with
	// json:"<name>" tags. Untagged fields stay default.
	wantPaths := []struct {
		name string
		a, b interface{}
	}{
		{"FileName", src.FileName, dst.FileName},
		{"FilePath", src.FilePath, dst.FilePath},
		{"Description", src.Description, dst.Description},
		{"ContentSize", src.ContentSize, dst.ContentSize},
		{"Content", src.Content, dst.Content},
		{"Address", src.Address, dst.Address},
		{"Directory", src.Directory, dst.Directory},
	}
	for _, tc := range wantPaths {
		if !reflect.DeepEqual(tc.a, tc.b) {
			t.Errorf("%s mismatch: src=%v dst=%v", tc.name, tc.a, tc.b)
		}
	}
}

// --- CanWrite --------------------------------------------------------------

func TestCanWrite_ReflectsUserCanWriteFlag(t *testing.T) {
	t.Parallel()
	fr := NewFileRecord("/", "f")
	if fr.CanWrite("anyone") {
		t.Error("CanWrite default true, expected false (UserCanWrite is false by default)")
	}
	fr.UserCanWrite = true
	if !fr.CanWrite("anyone") {
		t.Error("CanWrite false even with UserCanWrite=true")
	}
}
