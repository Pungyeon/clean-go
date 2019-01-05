package duplicates

import (
	"io/ioutil"
	"testing"
)

const (
	filehash = "2123251bdbfbb162fcd77b74f4954726461e8093"
)

func TestFileEntry(t *testing.T) {
	tt := []struct {
		name        string
		fullpath    string
		size        int64
		expectError bool
	}{
		{"handle existing file", "../testdata/text.txt", 100, false},
		{"handle non existing file", "../testdata/does_not_exist.txt", 100, true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			fileEntry := FileEntry{
				fullpath: tc.fullpath,
				size:     tc.size,
			}
			err := fileEntry.Handle(NewDuplicateIndex())
			if err != nil && tc.expectError == false {
				t.Errorf("expected error: %v, actual error: %v", err, tc.expectError)
			}
		})
	}
}

func TestFileHash(t *testing.T) {
	fileEntry := FileEntry{
		fullpath: "../testdata/text.txt",
		size:     100,
	}

	hash, err := fileEntry.newHash()
	if err != nil {
		t.Error(err)
	}

	if hash != filehash {
		t.Error(hash)
	}
}

func TestNilEntry(t *testing.T) {
	nilEntry := NilEntry{}

	result := nilEntry.Handle(&DuplicateIndex{})
	if result != nil {
		t.Error("nil entry should always return nil on handle, but instead return: " + result.Error())
	}
}

func TestEntryHandlers(t *testing.T) {
	entries, err := ioutil.ReadDir("../testdata")
	if err != nil {
		t.Fatal(err)
	}
	index := NewDuplicateIndex()

	for _, entry := range entries {
		if err := NewEntryHandler(entry, "../testdata").Handle(index); err != nil {
			t.Error(err)
		}
	}
}
