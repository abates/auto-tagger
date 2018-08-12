package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/abates/tag"
	"github.com/disintegration/imaging"
)

type Mp4 struct{}

func (mp4 *Mp4) Update(filename string, metadata Metadata) error {
	coverFile, err := ioutil.TempFile(path.Dir(filename), "")
	if err == nil {
		defer os.Remove(coverFile.Name())
		err = jpeg.Encode(coverFile, imaging.Resize(metadata.Poster(), 600, 0, imaging.Lanczos), &jpeg.Options{100})
		if err == nil {
			coverFile.Close()
		}
	}

	outputFile, err := ioutil.TempFile(path.Dir(filename), "")
	if err == nil {
		outputFile.Close()
	}

	if err == nil {
		args := []string{
			"-i", filename,
			"-i", coverFile.Name(),
			"-disposition:v:1", "attached_pic",
			"-map", "0", "-map", "1",
			"-metadata", fmt.Sprintf("title=%q", metadata.Title()),
			"-metadata", fmt.Sprintf("year=\"%d\"", metadata.Year()),
			"-metadata", fmt.Sprintf("synopsis=%q", metadata.Synopsis()),
			"-c", "copy",
			"-f", "mp4",
			"-y",
			outputFile.Name(),
		}

		cmd := exec.Command("ffmpeg", args...)
		err = cmd.Run()
	}

	if err == nil {
		err = os.Rename(outputFile.Name(), filename)
	}

	if _, err := os.Stat(outputFile.Name()); err == nil {
		os.Remove(outputFile.Name())
	}
	return err
}

func (mp4 *Mp4) Read(filename string) (Metadata, error) {
	metadata := &BaseMetadata{}
	file, err := os.Open(filename)
	if err == nil {
		m, err := tag.ReadFrom(file)
		if err == nil {
			tags := m.Raw()
			metadata.title = m.Title()
			if t, found := tags["ldes"]; found {
				metadata.synopsis = t.(string)
			}

			if t, found := tags["\xa9day"]; found {
				metadata.releaseDate, err = time.Parse("2006", t.(string))
			}

			if err == nil && m.Picture() != nil {
				metadata.poster, _, err = image.Decode(bytes.NewReader(m.Picture().Data))
			}
		}
	}
	return metadata, err
}
