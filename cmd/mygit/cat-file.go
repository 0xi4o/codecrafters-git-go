package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

func catFile(args []string) {
	switch flag := args[2]; flag {
	case "-p":
		path := fmt.Sprintf(".git/objects/%s/%s", args[3][0:2], args[3][2:])

		// read file
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
			os.Exit(1)
		}

		// decompress with zlib
		b := bytes.NewReader(data)
		r, err := zlib.NewReader(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error uncompressing file: %s\n", err)
			os.Exit(1)
		}
		defer r.Close()

		// read decompressed file content
		w, err := io.ReadAll(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading content: %s\n", err)
			os.Exit(1)
		}
		content := bytes.Split(w, []byte("\x00"))

		// print content
		fmt.Print(string(content[1]))

	default:
		fmt.Fprintf(os.Stderr, "Unknown flag for cat-file: %s\n", flag)
		os.Exit(1)
	}
}
