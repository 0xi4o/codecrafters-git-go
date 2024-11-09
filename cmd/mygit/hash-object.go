package main

import (
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// define type, enum, and method for the flags passed to the hash-object command
type Flag int

const (
	t Flag = iota
	w
)

// get the rune value of the flag
func (f Flag) Rune() rune {
	return [...]rune{'t', 'w'}[f]
}

// define type, enum, and methods for object type defined with -t flag
type ObjectType int

const (
	blob ObjectType = iota
	commit
	tag
	tree
)

// get the string value of the object type
func (t ObjectType) String() string {
	return [...]string{"blob", "commit", "tag", "tree"}[t]
}

// parse a string into an object type
func ParseObjectType(t string) ObjectType {
	types := make(map[string]ObjectType)
	for i := blob; i <= tree; i++ {
		types[i.String()] = i
	}
	objectType, ok := types[t]
	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid value for -t: %s. Possible values are blob, commit, tag, and tree.\n", t)
		os.Exit(1)
	}
	return objectType
}

// define a map for storing flags and corresponding values
type Flags map[string]interface{}

// parse arguments and update flags before executing the command
// note: this assumes that file name is the last argument
func parseArgs(args []string) Flags {
	flags := Flags{}
	for i, v := range args {
		if v[0] == '-' {
			switch rune(v[1]) {
			case t.Rune():
				flags[v] = ParseObjectType(args[i+1]).String()
			case w.Rune():
				flags[v] = true
			default:
				fmt.Fprintf(os.Stderr, "Unknown flag for hash-object: %s\n", v)
				os.Exit(1)
			}
		} else {
			flags["file"] = args[i]
		}
	}

	return flags
}

func hashObject(args []string) {
	flags := parseArgs(args[2:])
	objectType := blob.String()
	ot, typeFlagExists := flags["-t"]
	if typeFlagExists {
		objectType = ParseObjectType(ot.(string)).String()
	}

	writeToObjects, writeFlagExists := flags["-w"]
	filePath, filePathExists := flags["file"]
	if !filePathExists {
		fmt.Fprintf(os.Stderr, "Missing file name\n")
		os.Exit(1)
	}

	file, err := os.Open(filePath.(string))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		os.Exit(1)
	}

	fileStat, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting file size: %s\n", err)
		os.Exit(1)
	}

	fileSize := fileStat.Size()

	fileContent, err := io.ReadAll(io.Reader(file))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		os.Exit(1)
	}

	shaInput := fmt.Sprintf("%s %d\x00%s", objectType, fileSize, string(fileContent))
	h := sha1.New()
	h.Write([]byte(shaInput))
	sha := hex.EncodeToString(h.Sum(nil))
	fmt.Println(sha)

	if writeFlagExists && writeToObjects.(bool) {
		dirPath := fmt.Sprintf(".git/objects/%s", sha[0:2])
		path := fmt.Sprintf("%s/%s", dirPath, sha[2:])

		err := os.MkdirAll(dirPath, 0750)
		if err != nil && !os.IsExist(err) {
			fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			os.Exit(1)
		}

		file, err := os.Create(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating object file: %s\n", err)
			os.Exit(1)
		}
		defer file.Close()

		writer := zlib.NewWriter(file)
		_, err = writer.Write([]byte(shaInput))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing compressed data to file: %s\n", err)
			os.Exit(1)
		}

		err = writer.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error closing zlib writer: %s\n", err)
			os.Exit(1)
		}
	}
}
