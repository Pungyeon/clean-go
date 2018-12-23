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

Now all we need to do, is make a few adjustments to make entries and directory a single `type`. 