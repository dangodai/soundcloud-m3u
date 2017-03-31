package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type BandcampAlbum struct {
	Artist, Title string
	Tracks        []Track
}

// Determine whether URL is an album or an artist/label and
// call the appropriate function
func handleBandcampUrl(url string) {
	//albums and tracks seem to be handled the same
	if strings.Contains(url, "/album/") || strings.Contains(url, "/track/") {
		handleBandcampAlbum(url)
	} else if strings.Contains(url, ".bandcamp") {
		handleBandcampProfile(url)
	}
}

//Function to handle bandcamp profile/label urls
//Mostly just grabs the album URLs and calls handleBandcampUrl
func handleBandcampProfile(url string) {
	//Get the HTML body from the album page
	body, err := getHTMLBody(url)
	if err != nil {
		handleError(errors.New("bad bandcamp url"))
	}

	//Regex borrowed from the Soundscrape project: https://github.com/Miserlou/SoundScrape
	//Grabs all the album paths from a users profile
	albumregex := regexp.MustCompile(`<a href="(/(?:album|track)/[^>]+)">`)
	albums := albumregex.FindAllStringSubmatch(body, -1)

	//Call handleBandcampAlbum for each album
	for _, album := range albums {
		verboseMessage("Album found at: " + url + album[1])
		handleBandcampAlbum(url + album[1])
	}
}

//Function to handle a bandcamp album URL. It runs the process of gathering
//the information about the album and its tracks, and then saves them to
//a file
func handleBandcampAlbum(url string) {
	//Get the HTML body from the album page
	body, err := getHTMLBody(url)
	if err != nil {
		handleError(errors.New("bad bandcamp url"))
	}

	//Use regex to grab TralbumData from album page's javascript
	pattern := regexp.MustCompile("(?:var TralbumData = )((.|\n)*?)(?:;)")
	//Index 0 of the returned slice is the entire match, we just want
	//the first group, which excludes the var name and trailing semicolon
	//TODO: Error check (returns nil if there was no match)
	match := pattern.FindStringSubmatch(body)
	if len(match) < 1 {
		handleError(errors.New("no album found"))
		return
	}
	TralbumData := pattern.FindStringSubmatch(body)[1]

	//This regex should grab the Artist, Trackinfo, and "current" fields from the json
	//The rest of the json is too messy to use unmarshal, so we have to make our own
	//cleanpattern := regexp.MustCompile(`((?:artist: ")[A-Za-z]*(?:",))|(trackinfo: \[(.)*\])|(current: {.*})`)

	var album BandcampAlbum
	album.Artist = bandcampParseAlbumArtist(TralbumData)
	album.Tracks = bandcampParseAlbumTracks(TralbumData)
	album.Title = bandcampParseAlbumTitle(TralbumData)

	filename := album.Title + " by " + album.Artist
	writeTracksToFile(filename, album.Tracks)
}

func bandcampParseAlbumTitle(TralbumData string) string {
	titleregex := regexp.MustCompile(`(?:title":")(.*?)(?:")`)
	return titleregex.FindStringSubmatch(TralbumData)[1]
}

//Parse some data from the albums tracks, found in TralbumData
func bandcampParseAlbumTracks(TralbumData string) []Track {
	var tracks []Track

	//Use regex to grab the stream url, title, and duration for each tracks
	streamsregex := regexp.MustCompile(`(?:mp3-128":")(.*?)(?:")`)
	streams := streamsregex.FindAllStringSubmatch(TralbumData, -1)

	//TODO: Need better regex here. There is a "title" field for each track
	//but there is also title fields elsewhere in the json that I keep matching.
	//So we grab the text from "title_link" instead.
	titlesregex := regexp.MustCompile(`(?:title_link":"/track/)(.*?)(?:")`)
	titles := titlesregex.FindAllStringSubmatch(TralbumData, -1)

	durationsregex := regexp.MustCompile(`(?:duration":)(.*?)(?:,)`)
	durations := durationsregex.FindAllStringSubmatch(TralbumData, -1)

	//Make sure the regex found the same amount of tracks for each field
	//Otherwise give up
	count := len(streams)
	if len(titles) != count || len(durations) != count {
		handleError(errors.New("count mismatch"))
		return tracks
	}

	for i, _ := range streams {
		//convert duration to int
		duration, err := strconv.ParseFloat(durations[i][1], 64)
		if err != nil {
			duration = 0
		}

		stream := "https:" + streams[i][1]

		track := Track{Title: titles[i][1], Stream: stream, Duration: int(duration)}
		tracks = append(tracks, track)
	}

	return tracks
}

//Parse the album's artist from TralbumData
func bandcampParseAlbumArtist(TralbumData string) string {
	artistregex := regexp.MustCompile(`(?:artist: ")(.*?)(?:")`)
	artist := artistregex.FindStringSubmatch(TralbumData)
	if len(artist) > 1 {
		return artist[1]
	}
	return "Unknown Artist"
}

func getHTMLBody(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body), err
}
