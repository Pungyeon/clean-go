### Draft
# Cleaning Code in Go

## Introduction
So, this article is a little different from my others. Instead of focusing on a specific product, or solving a speficied problem, we will be looking at something a little more abstract. This article will be focusing on writing "clean code". The article will be starting with a short introduction as to what is defined by "clean code" and then, we will move onto a practical example, in which we refactor an example code snippet, into a cleaner version. You can find all code for this article at https://github.com/Pungyeon/clean-go

## What is "Clean Code"
The idea of clode code, is not something that is particularly rigid in definition. In my opinion, the closest thing to a defacto standard, are the books produced by Robert C. Martin (also known as "Uncle Bob"), who has written "Clean Code" and has produced an excellent video series on the topic. 

However, I will attempt to give a brief summary of what I believe to be clean code:

1. Easy to read code
	- Clean code is easy to read. In fact, it should be almost as easy to read as prose. If there is need for comments or the like, the code most likely isn't clean. It's intentions should be very clear, just from skimming the code.
2. Independant of rest of code base
	- Clean code ensures that if code changes in one part of the codebase, it shouldn't have to change the rest of the codebase. If we introduce a new database, we should be able to replace only the database logic and be able to test the code similarly with the new logic.
3. Testable
	- This is not only true for clean code, but for *all code*. It should be testable. If code is not testable, we can be very sure that it's not clean. 

There are many other additions to these sentiments. Code also shouldn't be duplicated, functions shouldn't be very long etc. However, we will cover this later. These three rules are, in my opinion, the most important aspects to writing clean code.

Whereas most aspects of clean code make sense and seems extremely intuitive, there are also some counterintuitive aspects of clean code. Writing clean code usually produces more lines of code than bad code (also referred to as sphagetti code). It's therefore very important to recognize, that writing clean code is not making the code "fat free" exclusively. The main goal of writing clean code is to make future development of code easier, and to reduce / eliminate bugs from programs. 

## Why "Clean Code"?


## Our Application
So, let's get right to it. I made a simple program, which traverses a file system and returns a list of duplicate files, based on their file contents. The way we are doing this is by reading the file, hashing the contents of the files as a `sha256` string, which is stored in a hash table, and then comparing the files on each traversal iteration. 

### Sphagetti Code
This is my first iteration of the program. Which was written pretty fast, without the consideration of anyone else going reading the code:

```go 
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
	if nbytes > 1024*1024*1024*1024 {
		return strconv.FormatInt(nbytes/(1024*1024*1024*1024), 10) + " TB"
	}
	if nbytes > 1024*1024*1024 {
		return strconv.FormatInt(nbytes/(1024*1024*1024), 10) + " GB"
	}
	if nbytes > 1024*1024 {
		return strconv.FormatInt(nbytes/(1024*1024), 10) + " MB"
	}
	if nbytes > 1024 {
		return strconv.FormatInt(nbytes/1024, 10) + " KB"
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
```

Going through the code via. the `main` method, we are parsing an input parameter `path`, and using this to read files from a directory. These files will be sent to the function `traverseDir`, in which we are also parsing two hash `map` objects `hashes` (all file hashes) and `duplicates` (all duplicate file hashes). Lastly, we are also inputting the `dupeSize` parameter, which will indicate the cummultative file size of our duplicate files. 

// explain traverseDir


## Refactoring

### Refactoring `toReadableSize`

First, we are going to be picking the low-hanging-fruits. The function `toReadableSize` looks pretty ugly. Firstly, we are using multiples of `1024`. For everyone who knows what this number represents, it makes sense, however, for anyone reading the code for the first time, this is just an ambiguous number. Therefore we will establish some global constants for the different integer values of the sizes that we are returning (GB, MB etc.). We use this when determining the readable size of `nbytes`, and change the if statement blocks into switch statements:

```go
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
```

Secondly, we are also getting rid of the `if` statements and converting to using a `switch` statement instead. We aren't doing much differently here, but it's a good example to start off with. The intention of this function is now much more clear, with very little effort.

### Refactoring `traverseDir`

Ok, let's go to the more interesting function, `traverseDir`. Why do we want to refactor this function? A good way to think about this, is to think of how you would describe this function in pseudio code and then compare it to your actual function. I'm thinking that this function could be reduced to the following pseudo-code.

```
traverseDir:
    for each entry in directory:
        if dir:
            return traverseDir
        if file:
            check file is duplicate
```

That is a lot less lines than what we have now... and definitely more readable than what we have now. The pseudo code is our target for what our function should look at. So, let's split it apart and see what we can do... 

Something that is nice about golangs, otherwise very criticised, error handling system, is that it's quite easy to spot when there is potential for refactoring. Whenever you see two `if err != nil` in the same function, you know you can split this out to a single function. In our case, this becomes:

```go
func traverseDir(...) {
    ...
    hash, err := newFileHash(fullpath)
    if err != nil {
        panic(err)
    }
    ...
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
```

Hm... we are still panicking in the case of an error, this is pretty dirty, so let's clean it up a little, by making `traverseDir` return an `error`, by adding `error` to the end of the function delcaration and replacing `panic(err)` with `return err`:

```go 
func traverseDir(hashes, duplicates map[string]string, dupeSize *int64, entries []os.FileInfo, directory string) error {
    ...
    return err
    ...
    return nil
}
```

Now, looking at the function signature, we can see that it's a bit... long? We are expecting five input parameters. It is considered best practice to have two (three at most), so there is no denying that this is a little too much.

So, how do you solve this? We need all the values, which is why we are passing them to the function. However, when we see this kind of pattern, it's usually a sign, that we need to extract a `type`. So, let's make a new `type`, which holds the paremeters that we need:

Here we create `DuplicateIndex` for keeping hold of our hashes and duplicates:
```go 
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
```

With this, we can actually replace `hashes`, `duplicates` and `dupeSize` from our function put parameters and also replace our insert of the hash:

```go
func traverseDir(index *DuplicateIndex, entries []os.FileInfo, directory string) {
    ...
    index.AddEntry(hash, fullpath, entry.Size())
    ...
}
```
But ah! We can actually make this a method on the `DuplicateIndex` `type` and that way, we now only have two input parameters. :clap We can also move the reading of the directory out of the for loop, and just accept a single parameter `path`. So now our method looks like this:

```go
func (index *DuplicateIndex) TraverseDirRecursively(directory string) error {
	entries, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fullpath := (path.Join(directory, entry.Name()))

		if entry.IsDir() {
			index.TraverseDirRecursively(fullpath)
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
```

So, we are almost happy now. However, the for lopp in `TraverseDirRecursively` is still smelling a little... what should we do? Well, one was we can get rid of the code smell, is that we get rid of the if statements inside, by using the strategy pattern. This means we will use a 'type builder', which will return the appropriate type determined by the input. This returned type will then have a `Handle` function, which will perform the appropriate action associated with the type. Let's see what this looks like in action.

We will need quite a lot of code, but please don't let that scare you off!

```go
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
	hash, err := newFileHash(entry.fullpath)
	if err != nil {
		return err
	}
	index.AddEntry(hash, entry.fullpath, entry.size)
	return nil
}

func (entry *NilEntry) Handle(index *DuplicateIndex) error {
	return nil
}
```

With these types, our `TraverseDirRecursively` can now be refactored to the following:

```go
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
```

It may seem extensive, to add almost 40 lines of code just to remove 14. However, there is a reason behind the madness. When looking at the `TraverseDirRecursively` function, it is now only 9 lines of code. This is very easily digestable by the brain where as 23 lines might be hard to contain at first. We are breaking down the functions, so that we can have our brain rest as much as possible, but also to make our tests 'rest' as much as possible. Now, testing is an absolute breeze. We can individually test every single scenario, very specifically. Which is great!

We can also attach the `newFileHash` function to our `FileEntry` `type` and remove the `path` input parameter. We can also rename this function. Since it's attached to our `FileEntry`, there is no need to specify that we are creating a `FileHash`:

```go
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
```

### Refactoring `main`

So, now we are pretty much all refactored on the functions of the program. All we need to do now, is to refactor the main function. 

Firstly, I don't like the `var err error` that has to go! Whenever we see this, it's a sign that we are doing something wrong :) 

The last thing we are going to do, is that we will create a `Result()` function on the `DuplicateIndex`, which will return a similar string to what we are printing now and rename our `toReadableString` function to `

```go
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
```

our final main function looks as such:

```go
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
```