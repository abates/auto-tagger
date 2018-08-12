package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"path"
	"time"

	"gopkg.in/h2non/filetype.v1"
)

type Metadata interface {
	Title() string
	Year() int
	Synopsis() string
	Poster() image.Image
}

type BaseMetadata struct {
	title       string
	releaseDate time.Time
	synopsis    string
	poster      image.Image
}

func (bm *BaseMetadata) Title() string       { return bm.title }
func (bm *BaseMetadata) Year() int           { return bm.releaseDate.Year() }
func (bm *BaseMetadata) Synopsis() string    { return bm.synopsis }
func (bm *BaseMetadata) Poster() image.Image { return bm.poster }

type Processor interface {
	Process(filename string) (Metadata, error)
	Match(filename string) bool
}

type Updater interface {
	Update(filename string, metadata Metadata) error
}

type Reader interface {
	Read(filename string) (Metadata, error)
}

var processors []Processor
var updaters map[string]Updater
var readers map[string]Reader

func processorLookup(filename string) (Processor, bool) {
	basename := path.Base(filename)
	for _, processor := range processors {
		if processor.Match(basename) {
			return processor, true
		}
	}
	return nil, false
}

func updaterLookup(filename string) (updater Updater, found bool) {
	t, err := filetype.MatchFile(filename)
	if err == nil {
		updater, found = updaters[t.MIME.Value]
	}
	return
}

func readerLookup(filename string) (reader Reader, found bool) {
	t, err := filetype.MatchFile(filename)
	if err == nil {
		reader, found = readers[t.MIME.Value]
	}
	return
}

func init() {
	processors = append(processors, &MovieProcessor{})

	updaters = make(map[string]Updater)
	updaters["video/mp4"] = &Mp4{}

	readers = make(map[string]Reader)
	readers["video/mp4"] = &Mp4{}
}

var skipExisting bool
var onlyLookupMovie bool

func processFile(filename string) (err error) {
	if processor, found := processorLookup(filename); found {
		if updater, found := updaterLookup(filename); found {
			if skipExisting {
				if reader, found := readerLookup(filename); found {
					metadata, err := reader.Read(filename)
					if err == nil {
						if metadata.Title() != "" {
							return fmt.Errorf("existing metadata")
						}
					}
				} else {
					err = fmt.Errorf("-s specified but no metadata reader found for %q", filename)
				}
			}
			metadata, err := processor.Process(filename)
			if err == nil {
				err = updater.Update(filename, metadata)
			}
		} else {
			err = fmt.Errorf("No updater found for %q", filename)
		}
	} else {
		err = fmt.Errorf("No processor found matching filename %q\n", filename)
	}

	return err
}

func lookupMovie(filename string) error {
	movie := &MovieProcessor{}
	metadata, err := movie.Process(filename)
	if err == nil {
		fmt.Printf("   Title: %v\n", metadata.Title())
		fmt.Printf("    Year: %v\n", metadata.Year())
		fmt.Printf("Synopsis: %v\n", metadata.Synopsis())
	}
	return err
}

func main() {
	flag.BoolVar(&skipExisting, "s", false, "skip updating files with existing metadata")
	flag.BoolVar(&onlyLookupMovie, "m", false, "lookup a movie by name (does not perform any file updates")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Printf("usage: %v file1 file2 file3 ...\n", os.Args[0])
		os.Exit(-1)
	}

	var err error
	for _, filename := range flag.Args() {
		if onlyLookupMovie {
			err = lookupMovie(filename)
		} else {
			err = processFile(filename)
		}

		if err == nil {
			log.Printf("Processed %q", filename)
		} else {
			log.Printf("%q: %v", filename, err)
		}
	}
}
