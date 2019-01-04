package main

import (
	"bytes"
	"crypto/sha1"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

const (
	TB = GB * 1000
	GB = MB * 1000
	MB = KB * 1000
	KB = 1000
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
	buf.WriteString(fmt.Sprintln("TOTAL DUPLICATE SIZE:", toReadableSize(index.dupeSize)))
	return buf.String()
}

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

func toReadableSize(nbytes int64) string {
	switch {
	case nbytes > TB:
		return strconv.FormatInt(nbytes/TB, 10) + " TB"
	case nbytes > GB:
		return strconv.FormatInt(nbytes/GB, 10) + " GB"
	case nbytes > MB:
		return strconv.FormatInt(nbytes/MB, 10) + " MB"
	case nbytes > KB:
		return strconv.FormatInt(nbytes/KB, 10) + " KB"
	}
	return strconv.FormatInt(nbytes, 10) + " B"
}

func main() {
	defaultPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dir := flag.String("path", defaultPath, "the path to traverse searching for duplicates")
	flag.Parse()

	index := NewDuplicateIndex()
	if err := index.TraverseDirRecursively(*dir); err != nil {
		panic(err)
	}

	fmt.Println(index.Result())
}
