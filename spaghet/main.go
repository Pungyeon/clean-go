package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync/atomic"
)

func traverseDir(hashes, duplicates map[string]string, dupeSize *int64, entries []os.FileInfo, directory string) {
	for _, entry := range entries {
		fullpath := (path.Join(directory, entry.Name()))

		if !entry.Mode().IsDir() && !entry.Mode().IsRegular() {
			continue
		}

		if entry.IsDir() {
			dirFiles, err := ioutil.ReadDir(fullpath)
			if err != nil {
				panic(err)
			}
			traverseDir(hashes, duplicates, dupeSize, dirFiles, fullpath)
			continue
		}
		file, err := ioutil.ReadFile(fullpath)
		if err != nil {
			panic(err)
		}
		hash := sha1.New()
		if _, err := hash.Write(file); err != nil {
			panic(err)
		}
		hashSum := hash.Sum(nil)
		hashString := fmt.Sprintf("%x", hashSum)
		if hashEntry, ok := hashes[hashString]; ok {
			duplicates[hashEntry] = fullpath
			atomic.AddInt64(dupeSize, entry.Size())
		} else {
			hashes[hashString] = fullpath
		}
	}
}

func toReadableSize(nbytes int64) string {
	if nbytes > 1000*1000*1000*1000 {
		return strconv.FormatInt(nbytes/(1000*1000*1000*1000), 10) + " TB"
	}
	if nbytes > 1000*1000*1000 {
		return strconv.FormatInt(nbytes/(1000*1000*1000), 10) + " GB"
	}
	if nbytes > 1000*1000 {
		return strconv.FormatInt(nbytes/(1000*1000), 10) + " MB"
	}
	if nbytes > 1000 {
		return strconv.FormatInt(nbytes/1000, 10) + " KB"
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

	traverseDir(hashes, duplicates, &dupeSize, entries, *dir)

	fmt.Println("DUPLICATES")
	for key, val := range duplicates {
		fmt.Printf("key: %s, val: %s\n", key, val)
	}
	fmt.Println("TOTAL FILES:", len(hashes))
	fmt.Println("DUPLICATES:", len(duplicates))
	fmt.Println("TOTAL DUPLICATE SIZE:", toReadableSize(dupeSize))
}

// running into problems of not being able to open directories inside .app folders
