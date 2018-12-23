### Draft

### Process of Refactoring

Create Tests (?) 

- Copy all code to the clean main package

Start Refactoring :

### Refactoring `toReadableSize`

First we make some global constants for the different integer values of the sizes that we are returning (GB, MB etc.). We use this when determining the readable size of `nbytes`, and change the if statement blocks into switch statements:

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

### Refactoring `traverseDir`

Ok, let's go to the more meatier function, traverseDir. Why do we want to refactor this function? Well, let's think about what our peudo code for this would look like:

```
traverseDir:
    for each entry in directory:
        if dir:
            return traverseDir
        if file:
            check file is duplicate
```

That is a lot more lines that we have now... and definitely more readable than what we have now. The pseudo code is our aim for what our function should look at. So, let's split it apart and see what we can do... 

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