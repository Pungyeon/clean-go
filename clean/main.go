package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Pungyeon/clean-go-code/clean/duplicates"
)

func main() {
	defaultPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dir := flag.String("path", defaultPath, "the path to traverse searching for duplicates")
	flag.Parse()

	index := duplicates.NewDuplicateIndex()
	if err := index.TraverseDirRecursively(*dir); err != nil {
		panic(err)
	}

	fmt.Println(index.Result())
}
