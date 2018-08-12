package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/ryanbradynd05/go-tmdb"
)

var TMDB_API_KEY = "54c287e0736e501f63506deb7afd25dd"
var TMDB_IMAGE_BASE_PATH = "http://image.tmdb.org/t/p"

var moviePattern = regexp.MustCompile(`^(?P<title>[^\(]+)\s\((?P<year>\d{4})\)\.(?P<extension>\w{3})$`)
var db *tmdb.TMDb

func mapFilename(re *regexp.Regexp, filename string) map[string]string {
	values := make(map[string]string)
	match := re.FindStringSubmatch(filename)
	for i, name := range re.SubexpNames() {
		values[name] = match[i]
	}
	return values
}

func downloadImage(path string) (img image.Image, err error) {
	resp, err := http.Get(TMDB_IMAGE_BASE_PATH + "/original/" + path)
	if err == nil {
		defer resp.Body.Close()
		if resp.Header.Get("Content-Type") == "image/jpeg" {
			img, err = jpeg.Decode(resp.Body)
		} else if strings.HasPrefix(resp.Header.Get("Content-Type"), "image") {
			img, _, err = image.Decode(resp.Body)
		} else {
			err = fmt.Errorf("unknown content type %v", resp.Header["Content-Type"])
		}
	}
	return
}

type MovieProcessor struct{}

func (mp *MovieProcessor) Lookup(title, year string) (movie *tmdb.Movie, err error) {
	results, err := db.SearchMovie(title, map[string]string{"year": year})
	if err == nil {
		if results.TotalResults == 1 {
			return db.GetMovieInfo(results.Results[0].ID, map[string]string{})
		} else {
			titles := make([]string, 0)
			for _, result := range results.Results {
				titles = append(titles, fmt.Sprintf("%q", result.Title))
				if result.Title == title {
					return db.GetMovieInfo(result.ID, map[string]string{})
				}
			}
			err = fmt.Errorf("Multiple results found: %v", strings.Join(titles, ","))
		}
	}
	return
}

func (mp *MovieProcessor) Process(filename string) (Metadata, error) {
	metadata := &BaseMetadata{}
	name := path.Base(filename)
	values := mapFilename(moviePattern, name)
	movie, err := mp.Lookup(values["title"], values["year"])
	if err == nil {
		metadata.title = movie.Title
		metadata.synopsis = movie.Overview
		metadata.releaseDate, err = time.Parse("2006-01-02", movie.ReleaseDate)

		if err == nil {
			metadata.poster, err = downloadImage(movie.PosterPath)
		}
	}
	return metadata, err
}

func (mp *MovieProcessor) Match(name string) bool {
	return moviePattern.Match([]byte(name))
}

func init() {
	config := tmdb.Config{
		ApiKey: TMDB_API_KEY,
	}

	db = tmdb.Init(config)
}
