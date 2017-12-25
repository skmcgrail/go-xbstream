package main

import (
	"github.com/skmcgrail/go-xbstream/xbstream"
	"io"
	"log"
	"os"
	"sync"
)

func init() {
	log.SetOutput(os.Stderr)
}

func main() {
	w := xbstream.NewWriter(os.Stdout)

	wg := sync.WaitGroup{}

	for _, f := range os.Args[1:] {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			b := make([]byte, xbstream.MinimumChunkSize)

			if file, err := os.Open(path); err == nil {
				fw, err := w.Create(path)
				if err != nil {
					log.Fatal(err)
				}

				for {
					n, err := file.Read(b)
					if err != nil {
						if err == io.EOF {
							break
						}
						log.Fatal(err)
					}
					if _, err := fw.Write(b[:n]); err != nil {
						log.Fatal(err)
						break
					}
				}

				err = fw.Close()
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Printf("unable to open file %s", file)
			}
		}(f)
	}

	wg.Wait()

	err := w.Close()
	if err != nil {
		log.Fatal(err)
	}
}
