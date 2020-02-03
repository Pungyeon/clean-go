# Cleaning Code in Go

## Introduction
So, this article is a little different from my others. Instead of focusing on a specific product, or solving a speficied problem, we will be looking at something a little more abstract. This article will be focusing on writing "clean code" with golang. The article will be starting with a short introduction as to what is defined by "clean code" and then, we will move onto a practical example, in which we refactor an example application, into a cleaner version. You can find all code for this article at https://github.com/Pungyeon/clean-go

## What is "Clean Code"
The idea of clean code, is not something that is particularly rigid in definition. In my opinion, the closest thing to a defacto standard, are the books produced by Robert C. Martin (also known as "Uncle Bob"), who has written the "Clean Code" series, as well as having produced an excellent and extensive video series on the topic. 

However, I will attempt to give a brief summary of what I believe to be clean code:

1. Easy to read code
	- Clean code is easy to read. In fact, it should be almost as easy to read as prose. If there is need for comments or the like, the code most likely isn't clean. It's intentions should be very clear, just from skimming the code.
2. Independent of rest of code base
	- Clean code ensures that if code changes in one part of the codebase, the rest of the codebase is essentially unaffected. In other words, code is segregated into functionality silos, independent of the rest of the code base.
3. Testable
	- If code is not testable, we can be very sure that it's not clean. Of course, *all code* should be tested, this is not necessarily something that is strictly related to clean code. Making code testable, however, is a big aspect of clean code.

There are many other additions to these sentiments. Code shouldn't be duplicated, functions shouldn't be very long etc. However, we will cover this later. These three rules are, in my opinion, the most important aspects to writing clean code.

Whereas most aspects of clean code make sense and seems extremely intuitive, there are also some counterintuitive aspects of clean code. Writing clean code can potentially produce more lines of code than dirty code (also referred to as smelly or sphagetti code). It's therefore very important to recognize, that writing clean code is not making the code "fat free" exclusively. The main goal of writing clean code is to make future development of code easier, and to reduce / eliminate introdution of bugs to applications.

> NOTE: In this article, I will not be writing tests along with the refactoring. Writing tests before refactoring (and before developing for that matter), is extremely important when writing clean code. However, I typically find that explaining TDD in text, rather than in video is not enjoyable for the writer, nor the reader. However, please please please, write tests when refactoring, to ensure that your refactoring is not destroying your code. I have provided some test examples in the source code for this article.

## Our Application
So, let's get right to it. I made a simple program, which traverses a file system and returns a list of duplicate files, based on their file contents. The way we are doing this is by reading the file and hashing the contents as a `sha256` string, which is stored in a hash table, and then comparing the files on each traversal iteration. 

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
```

Going through the code via. the `main` method, we are parsing an input parameter `path`, and using this to read files from a directory. These files will be sent to the function `traverseDir`, in which we are also parsing two hash `map` objects `hashes` (all file hashes) and `duplicates` (all duplicate file hashes). Lastly, we are also inputting the `dupeSize` parameter, which will indicate the cummultative file size of our duplicate files. 

Finally, we print out our results in a 'human readable' format. Instead of presenting our results as byte count, we will convert them to the appropriate size unit (KB, MB, GB etc.).

## Refactoring

### Refactoring `toReadableSize`

First, we are going to be picking the low-hanging-fruits. The function `toReadableSize` looks pretty ugly. Firstly, we are using multiples of `1000`. For everyone who knows what this number represents, it makes sense, however, for anyone reading the code for the first time, this number is rather ambiguous. Therefore, we will establish some global constants for the different values of the sizes that we are returning (GB, MB etc.). We use this when determining the readable size of `nbytes`, and change the if statement blocks into switch statements. As you might have noticed, we are only returning integers, where it would make more sense to return floats:

```go
const (
	TB = GB * 1000.0
	GB = MB * 1000.0
	MB = KB * 1000.0
	KB = 1000.0
)


func ToReadableSize(nbytes int64) string {
	switch {
	case nbytes > TB:
		return strconv.FormatFloat(float64(nbytes)/TB, 'f', 2, 64) + " TB"
	case nbytes > GB:
		return strconv.FormatFloat(float64(nbytes)/GB, 'f', 2, 64) + " GB"
	case nbytes > MB:
		return strconv.FormatFloat(float64(nbytes)/MB, 'f', 2, 64) + " MB"
	case nbytes > KB:
		return strconv.FormatFloat(float64(nbytes)/KB, 'f', 2, 64) + " KB"
	}
	return strconv.FormatFloat(float64(nbytes), 'f', 2, 64) + " B"
}
```

However, this is still very ugly and just as (if not more unreadable) than before. There is a lot of code duplication here, which we should get rid of. So let's make our own `toFloatString` function:

```go
func toFloatString(nbytes int64, divider float64) string {
	return strconv.FormatFloat(float64(nbytes)/divider, 'f', 2, 64)
}

func ToReadableSize(nbytes int64) string {
	switch {
	case nbytes > TB:
		return toFloatString(nbytes, TB) + " TB"
	case nbytes > GB:
		return toFloatString(nbytes, GB) + " GB"
	case nbytes > MB:
		return toFloatString(nbytes, MB) + " MB"
	case nbytes > KB:
		return toFloatString(nbytes, KB) + " KB"
	}
	return strconv.FormatInt(nbytes, 10) + " B"
}
```

Now, our function is nice and readable again. This refactor obviously isn't game changing, but it's a good example to start off with. The intention of this function is now much clearer, with very little effort.

### Refactoring `traverseDir`

Ok, let's go to the more interesting function, `traverseDir`. Why do we want to refactor this function? A good way to think about this, is to think of how you would describe this function in pseudo code and then compare it to your actual code. I'm thinking that this function could be reduced to the following pseudo-code.

```
traverseDir:
    for each entry in directory:
        if dir:
            return traverseDir
        if file:
            check file is duplicate
```

That is a lot less lines than what we have now... and definitely more readable than what we have now. Pseudo code is a pretty good way to establish a 'goal' for what your clean code should look like. At the very least, you should aim to make your actual code as readable as pseudo code. We can do this, by moving code into functions with descriptive names. This however, is an iterative process. We will start small and bit by bit, we will find a solution as to how to make our code simple and readable.

So let's look for code, which we can move out of this function... Something that is nice about golang's, otherwise very criticised, error handling system, is that it's quite easy to spot when there is potential for refactoring. Whenever you see two `if err != nil` statements in the same function, you know you can split this out to a single function. In our case, this:

```go
func traverseDir(...)
	...
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
	...
}
```

Can be refactored to the following:

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

The justification behind this, is that when reading our `traverseDir` we aren't immediately concerned with how we are creating a new file hash (sum). We just need to know that we are creating a new file hash. If we want to dig into the details of this, then we can by looking at our `newFileHash`. In other words, we are removing unecessary clutter from the function, improving readability.

Looking for more low-hanging fruites, we are still panicking in the case of an error, this is pretty dirty, so let's clean it up a little, by making `traverseDir` return an `error`, by adding `error` to the end of the function delcaration and replacing `panic(err)` with `return err`:

```go 
func traverseDir(hashes, duplicates map[string]string, dupeSize *int64, entries []os.FileInfo, directory string) error {
    ...
    return err
    ...
    return nil
}
```

Now, looking at the function signature, we can see that it's a bit... long? We are expecting five input parameters. Not only does this make our function signature super long, it can also makes it very confusing to read on invokation. Consider the following code (taken from the golang rabbitmq tutorial):

```go
q, err := ch.QueueDeclare("hello",false,false,false,false,nil)
```
There is absolutely no chance of understanding what this means. We know that we are declaring a queue, but all the boolean inputs... well, they could be anything? So, we have to either look at the source code or look at the documentation. This is tedious and slows down development speed and increases the risk of mistakes. A good rule of thumb is to have two input parameters (three at most), to try to avoid this type of confusion.

Generally, if there are more input parameters it is recommended to extract a type (creating a new type, which will be used as the input parameters). As an example:

```go
type QueueOptions struct {
  Name 	string
  Durable bool
  DeleteWhenUsed bool
  Exclusive bool
  NoWait bool
  Arguments interface{}
}
```

Now, our declaration of our queue, could look something like the following:

```go
q, err := ch.NewQueue(QueueOptions{
	Name: "hello",
	Durable: true,
	DeleteWhenUsed: false,
	Exclusive: false,
	NoWait: false,
	Arguments: nil,
})
```

Now there is, at the very least, less confusion as to what kind of queue that we are declaring. We can very easily identify that our queue name is `hello` and is a `durable` queue. Another way to go about this, is to create a wrapper function, which explains the type of queue we are creating. This is preferable, when you have no control over the code, such as when using a library:

```go
func DeclareDurableQueue() (ch.Queue, error) {
	return ch.QueueDeclare("hello", true, false, false, false, nil)
}
```

So, how do we go about solving this issue for our `traverseDir`? We need all the values, which is why we are passing them to the function. However, when we see this kind of pattern, it's usually a sign, that we need to extract a `type`. So, let's make a new `type`, which holds the paremeters that we need:

Here we create `DuplicateIndex` for keeping track of our hashes and duplicates:

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

With this, we can actually replace `hashes`, `duplicates` and `dupeSize` from our function parameters and also replace our insert of the hash:

```go
func traverseDir(index *DuplicateIndex, entries []os.FileInfo, directory string) {
    ...
    index.AddEntry(hash, fullpath, entry.Size())
    ...
}
```

But ah! We can actually make this a method on the `DuplicateIndex` `type` and that way, we now only have two input parameters :clap: We can also move the reading of the directory out of the for loop, and just accept a single parameter `path`. So now our method looks like this:

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

> Notice that we have also renamed our function name for clarity

Now we are almost happy. However, the for lopp in `TraverseDirRecursively` is still smelling a little... what should we do? Well, one way we can get rid of the code smell, is to get rid of the `if` statements inside, by creating an `interface` together with a factory-like constructor. This means we will return the appropriate type determined by the input of the constructor. This returned type will implement an `interface`, which implements a single function: `Handle`. This function will perform the appropriate action associated with the type. Let's see what this looks like in action.

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

It may seem extensive, to add almost 40 lines of code just to remove 14. However, there is a reason behind the madness. When looking at the `TraverseDirRecursively` function, it is now only 9 lines of code. This is very easily digestable by the brain, whereas 23 lines might be hard to contain at first. The big gain though, is that we are isolating code, we can test all of our functions very easily and understand exactly what they do, with very little effort. Another great advantage of this isolation, is that we are also making our `TraverseDirRecursively` more dynamic. If we find out, that there is a new type of entry that we need to handle (Shortcut for example), we can just add a new type implementing `EntryHandler` and add it to our mini-factory `NewEntryHandler`. We are now <b>only</b> changing the logic of `NewEntryHandler` as every other code addition, is completely separate. The obvious advantage of this is, it makes it easier to implement new code, without it breaking the rest of our code. We like this :thumbs_up:

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

Firstly, I don't like the `var err error` that has to go! Whenever we see this, it's a sign that we are doing something wrong (in my opinion :)) Normally, this indicates that we should move our code into a new function, but in this case, we can actually just move the logic around a little...

The last thing we are going to do, is that we will create a `Result()` function on the `DuplicateIndex`, which will return a similar string to what we are printing now:

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

this makes our final `main` function, look like this:

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

We could do some more refactoring, but for this short article, I think this is a good point to stop. Of course, in actual code, we would separate these functions into packages, to separate / isolate the responsibility of the code. However, again for the brevity of this article, I have decided to omit this refactoring step. You can, however, see how I decided to do this in the source code. 

Now, let's sum up the result of our code refactor:
* Our code is now easy to implement for other developers. 
* It's much easier to read than before. We can skim the code to begin with, and then go into detail on the parts that we wish to. There is less ambiguous / vague code, making everything generally easier to comprehend.
* Our code is super easy to test. This makes further development a lot easier and decreases the chances for bugs, for this very reason.

As mentioned to begin with 'clean code' is not necessarily super well defined and sometimes comes down to subjective opinion on what is 'more readable' or 'nicer looking'. However, I hope this article gave some insight as to why it's important to refactor your code, as well as how easy it actually is!

Let me know if you have any feedback or questions on this articles content, by sending me an e-mail at lasse@jakobsen.dev thanks! :wave:
