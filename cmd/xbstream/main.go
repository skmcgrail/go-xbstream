package main

import (
	"github.com/akamensky/argparse"
	"github.com/skmcgrail/go-xbstream/xbstream"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func init() {
	log.SetOutput(os.Stderr)
}

func main() {
	parser := argparse.NewParser("xbstream", "Go implementation of the xbstream archive format")

	createCmd := parser.NewCommand("create", "create xbstream archive")
	createFile := createCmd.File("o", "output", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666, &argparse.Options{})
	createList := createCmd.List("i", "input", &argparse.Options{Required: true})

	extractCmd := parser.NewCommand("extract", "extract xbstream archive")
	extractFile := extractCmd.File("i", "input", os.O_RDONLY, 0600, &argparse.Options{})
	extractOut := extractCmd.String("o", "output", &argparse.Options{})

	if err := parser.Parse(os.Args); err != nil {
		log.Fatal(err)
	}

	if createCmd.Happened() {
		writeStream(createFile, createList)
	} else if extractCmd.Happened() {
		readStream(extractFile, *extractOut)
	}
}

func readStream(file *os.File, output string) {
	var err error

	if *file == (os.File{}) {
		file = os.Stdin
	}

	if output == "" {
		output, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	if err = os.MkdirAll(output, 0777); err != nil {
		log.Fatal(err)
	}

	r := xbstream.NewReader(file)

	files := make(map[string]*os.File)

	var f *os.File
	var ok bool

	for {
		chunk, err := r.Next()
		if err != nil {
			log.Fatal(err)
			break
		}

		if chunk.Type == xbstream.ChunkTypeEOF {
			break
		}

		fPath := string(chunk.Path)

		if f, ok = files[fPath]; !ok {
			newFPath := filepath.Join(output, fPath)
			if err = os.MkdirAll(filepath.Dir(newFPath), 0777); err != nil {
				break
				log.Fatal(err)
			}

			f, err = os.OpenFile(newFPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
			if err != nil {
				log.Fatal(err)
				break
			}
			files[fPath] = f
		}

		f.Seek(int64(chunk.PayOffset), io.SeekStart)
		if _, err = io.Copy(f, chunk); err != nil {
			log.Fatal(err)
			break
		}
	}

	for _, v := range files {
		if err = v.Close(); err != nil {
			log.Fatal(err)
			break
		}
	}
}

func writeStream(file *os.File, input *[]string) {
	if *file == (os.File{}) {
		file = os.Stdout
	}

	w := xbstream.NewWriter(file)

	wg := sync.WaitGroup{}

	for _, f := range *input {
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
