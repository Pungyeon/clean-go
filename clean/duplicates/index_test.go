package duplicates

import "testing"

const (
	result = `DUPLICATES
key: ../testdata/copy.txt, val: ../testdata/text.txt
TOTAL FILES: 2
DUPLICATES: 1
TOTAL DUPLICATE SIZE: 41 B
`
)

func TestTraverseDir(t *testing.T) {
	tt := []struct {
		name        string
		directory   string
		expectError bool
	}{
		{"traverse existing directory", "../testdata", false},
		{"traverse non-existing directory", "../does_not_exist", true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := NewDuplicateIndex().TraverseDirRecursively(tc.directory)
			if err != nil && tc.expectError == false {
				t.Error(err)
			}
		})
	}
}

func TestTraverseDirResult(t *testing.T) {
	index := NewDuplicateIndex()
	if err := index.TraverseDirRecursively("../testdata"); err != nil {
		t.Error(err)
	}
	if index.Result() != result {
		t.Error("unexpected result")
		t.Error(index.Result())
	}
}
