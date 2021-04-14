package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, root string, printFiles bool) error {
	return dirTreeInternal(out, root, printFiles, "")
}

func dirTreeInternal(out io.Writer, root string, printFiles bool, prefix string) error {
	entries, err := ioutil.ReadDir(root)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	var prev os.FileInfo

	for _, entry := range entries {
		if !printFiles && !entry.IsDir() {
			continue
		}
		if prev != nil {
			print(out, prev, prefix, false)
			dirTreeInternal(out, root+string(os.PathSeparator)+prev.Name(), printFiles, prefix+"│	")
		}
		prev = entry
	}

	if prev != nil {
		print(out, prev, prefix, true)
		dirTreeInternal(out, root+string(os.PathSeparator)+prev.Name(), printFiles, prefix+"	")
	}

	return nil
}

func print(out io.Writer, entry os.FileInfo, prefix string, last bool) {
	var entryStr string
	if entry.IsDir() {
		entryStr = entry.Name()
	} else {
		var sizeStr string
		if entry.Size() == 0 {
			sizeStr = " (empty)"
		} else {
			sizeStr = " (" + fmt.Sprint(entry.Size()) + "b)"
		}
		entryStr = entry.Name() + sizeStr
	}

	if last {
		fmt.Fprint(out, prefix, "└───", entryStr, "\n")
		return
	}
	fmt.Fprint(out, prefix, "├───", entryStr, "\n")
}
