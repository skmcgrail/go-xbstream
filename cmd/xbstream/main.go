package main

import (
	"github.com/skmcgrail/go-xbstream/xbstream"
	"io"
	"log"
	"os"
)

func init() {
	log.SetOutput(os.Stderr)
}

func main() {
	w := xbstream.NewWriter(os.Stdout)

	b := make([]byte, xbstream.MinimumChunkSize)

	for _, f := range os.Args[1:] {
		if file, err := os.Open(f); err == nil {
			fw, err := w.Create(f)
			if err != nil {
				continue
			}

			for {
				n, err := file.Read(b)
				log.Println(n)
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Fatal(err)
					break
				}
				if _, err := fw.Write(b[:n]); err != nil {
					log.Println(err)
					break
				}
			}

			err = fw.Close()
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Printf("unable to open file %s", file)
		}
	}

	err := w.Close()
	if err != nil {
		log.Println(err)
	}
}
