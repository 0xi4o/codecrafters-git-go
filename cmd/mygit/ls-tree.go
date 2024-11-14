package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strings"
)

func lsTree(args []string) {
	switch flag := args[2]; flag {
	case "--name-only":
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

		header, content, _ := bytes.Cut(w, []byte("\x00"))

		t, _, _ := strings.Cut(string(header), " ")

		if t == "tree" {
			remaining := string(content)
			// lsTree := make(map[string]interface{})
			for remaining != "" {
				_, rest, _ := strings.Cut(remaining, " ")
				filename, rest, _ := strings.Cut(rest, "\x00")
				r := strings.NewReader(rest)
				buf := make([]byte, 20)
				if _, err := io.ReadAtLeast(r, buf, 20); err != nil {
					fmt.Println("error: ", err)
					break
				}
				// sha := make([]byte, hex.EncodedLen(len(buf)))
				// hex.Encode(sha, buf)
				// fmt.Println(string(sha))
				remaining = strings.Replace(rest, string(buf), "", 1)
				fmt.Println(filename)
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown flag for cat-file: %s\n", flag)
		os.Exit(1)
	}
}
