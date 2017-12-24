package main

import (
	"github.com/skmcgrail/go-xbstream/xbstream"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	w := xbstream.NewWriter(os.Stdout)

	for _, f := range os.Args[1:] {
		if file, err := os.Open(f); err == nil {
			path := filepath.Base(f)
			fw, err := w.Create(path)
			if err != nil {
				continue
			}

			_, err = io.Copy(fw, file)
			if err != nil {
				log.Fatal(err)
			}

			fw.Close()
		} else {
			log.Printf("unable to open file %s", file)
		}
	}

	err := w.Close()
	if err != nil {
		log.Fatal(err)
	}
}
