package duplicates

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type EntryHandler interface {
	Handle(*DuplicateIndex) error
}

type DirEntry struct {
	fullpath string
}

type FileEntry struct {
	fullpath string
	size     int64
}

type NilEntry struct{}

func NewEntryHandler(entry os.FileInfo, directory string) EntryHandler {
	fullpath := path.Join(directory, entry.Name())
	if entry.Mode().IsDir() {
		return &DirEntry{fullpath}
	}
	if entry.Mode().IsRegular() {
		return &FileEntry{fullpath, entry.Size()}
	}
	return &NilEntry{}
}

func (entry *DirEntry) Handle(index *DuplicateIndex) error {
	return index.TraverseDirRecursively(entry.fullpath)
}

func (entry *FileEntry) Handle(index *DuplicateIndex) error {
	hash, err := entry.newHash()
	if err != nil {
		return err
	}
	index.AddEntry(hash, entry.fullpath, entry.size)
	return nil
}

func (entry *FileEntry) newHash() (string, error) {
	file, err := ioutil.ReadFile(entry.fullpath)
	if err != nil {
		return "", err
	}
	hash := sha1.New()
	if _, err := hash.Write(file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (entry *NilEntry) Handle(index *DuplicateIndex) error {
	return nil
}
