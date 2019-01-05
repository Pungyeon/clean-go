package duplicates

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/Pungyeon/clean-go-code/clean/utils"
)

type DuplicateIndex struct {
	hashes     map[string]string
	duplicates map[string]string
	dupeSize   int64
}

func NewDuplicateIndex() *DuplicateIndex {
	return &DuplicateIndex{
		hashes:     map[string]string{},
		duplicates: map[string]string{},
	}
}

func (index *DuplicateIndex) AddEntry(hash, path string, size int64) {
	if entry, ok := index.hashes[hash]; ok {
		index.duplicates[entry] = path
		index.dupeSize += size
		return
	}
	index.hashes[hash] = path
}

func (index *DuplicateIndex) TraverseDirRecursively(directory string) error {
	entries, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := NewEntryHandler(entry, directory).Handle(index); err != nil {
			return err
		}
	}
	return nil
}

func (index *DuplicateIndex) Result() string {
	buf := &bytes.Buffer{}
	buf.WriteString("DUPLICATES\n")
	for key, val := range index.duplicates {
		buf.WriteString(
			fmt.Sprintf("key: %s, val: %s\n", key, val),
		)
	}
	buf.WriteString(fmt.Sprintln("TOTAL FILES:", len(index.hashes)))
	buf.WriteString(fmt.Sprintln("DUPLICATES:", len(index.duplicates)))
	buf.WriteString(fmt.Sprintln("TOTAL DUPLICATE SIZE:", utils.ToReadableSize(index.dupeSize)))
	return buf.String()
}
