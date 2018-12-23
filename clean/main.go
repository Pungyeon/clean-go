package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
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

func traverseDir(index *DuplicateIndex, entries []os.FileInfo, directory string) error {
	for _, entry := range entries {
		fullpath := (path.Join(directory, entry.Name()))

		if entry.IsDir() {
			dirFiles, err := ioutil.ReadDir(fullpath)
			if err != nil {
				return err
			}
			traverseDir(index, dirFiles, fullpath)
			continue
		}
		if !entry.Mode().IsRegular() {
			continue
		}

		hash, err := newFileHash(fullpath)
		if err != nil {
			return err
		}
		index.AddEntry(hash, fullpath, entry.Size())
	}
	return nil
}

func newFileHash(path string) (string, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha1.New()
	if _, err := hash.Write(file); err != nil {
		return "", err
	}
	hashSum := hash.Sum(nil)
	return fmt.Sprintf("%x", hashSum), nil
}

const TB = 1099511627776
const GB = 1073741824
const MB = 1048576
const KB = 1024

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
	var err error
	dir := flag.String("path", "", "the path to traverse searching for duplicates")
	flag.Parse()

	if *dir == "" {
		*dir, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}

	hashes := map[string]string{}
	duplicates := map[string]string{}
	var dupeSize int64

	entries, err := ioutil.ReadDir(*dir)
	if err != nil {
		panic(err)
	}

	if err := traverseDir(hashes, duplicates, &dupeSize, entries, *dir); err != nil {
		panic(err)
	}

	fmt.Println("DUPLICATES")
	for key, val := range duplicates {
		fmt.Printf("key: %s, val: %s\n", key, val)
	}
	fmt.Println("TOTAL FILES:", len(hashes))
	fmt.Println("DUPLICATES:", len(duplicates))
	fmt.Println("TOTAL DUPLICATE SIZE:", toReadableSize(dupeSize))
}

// running into problems of not being able to open directories inside .app folders
