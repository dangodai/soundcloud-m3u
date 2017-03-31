package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kennygrant/sanitize"
	"github.com/yanatan16/golang-soundcloud/soundcloud"
)

type Track struct {
	//The http stream URL
	Stream   string
	Duration int
	Title    string
	Artist   string
}

var (
	ClientID                          string //Client ID from soundcloud api
	PlaylistDir                       string //Directory to save m3u playlists to
	URL                               string //Soundcloud URL the grab resource from
	favouritesb, playlistsb, verboseb bool
	api                               *soundcloud.Api
)

func init() {
	flag.StringVar(&ClientID, "id", "2t9loNQH90kzJcsFCODdigxfp325aq4z", "ClientID to use for Soundcloud API")
	flag.StringVar(&PlaylistDir, "d", "", "Directory to save generated playlists to (shorthand)")
	flag.StringVar(&PlaylistDir, "dir", "", "Directory to save generated playlists to")
	flag.StringVar(&URL, "u", "", "Soundcloud URL to generate playlists from (shorthand)")
	flag.StringVar(&URL, "url", "", "Soundcloud URL to generate playlists from")
	flag.BoolVar(&verboseb, "v", false, "Verbose logging")
	flag.BoolVar(&verboseb, "verbose", false, "Verbose logging")
	flag.BoolVar(&favouritesb, "f", false, "If set, a profile URL will generate a playlist of that user's favourited tracks (shorthand)")
	flag.BoolVar(&favouritesb, "favourites", false, "If set, a profile URL will generate a playlist of that user's favourited tracks")
	flag.BoolVar(&playlistsb, "s", false, "If set, a profile URL will generate a playlist of that user's sets (playlists) (shorthand)")
	flag.BoolVar(&playlistsb, "sets", false, "If set, a profile URL will generate a playlist of that user's sets (playlists)")
	flag.Parse()
	api = &soundcloud.Api{ClientId: ClientID}
}

func main() {
	if URL == "" {
		handleError(errors.New("no url"))
	} else if strings.Contains(URL, "soundcloud.com") {
		handleSoundcloudUrl(URL)
	} else if strings.Contains(URL, "bandcamp.com") {
		handleBandcampUrl(URL)
	}
}

func handleSoundcloudUrl(url string) {
	//Resolve the given URL to find resource and ID
	resource, id, err := Resolve(url)
	if err != nil {
		handleError(err)
		return
	}

	//Call appropriate function for requested resource
	switch resource {
	case "tracks":
		verboseMessage("Track URL received")
		fromTrack(id)
	case "playlists":
		verboseMessage("Playlist URL received")
		fromPlaylist(id)
	case "users":
		verboseMessage("User URL received")
		fromUser(id)
	default:
		handleError(errors.New("unknown resource"))
	}

}

//Save m3u from a track url (just a single track)
func fromTrack(id uint64) {
	track, err := api.Track(id).Get(nil)
	if err != nil {
		panic(err)
	}

	tracks := make([]*soundcloud.Track, 1)
	tracks[0] = track

	filename := fmt.Sprintf("(Track) %v - %v", track.Title, track.SubUser.User.Username)
	writeTracksToFile(filename, soundcloudTracksToGenericTracks(tracks))
}

// fromUser saves an m3u file containing the given user's uploads.
// Optionally, if the -f flag is set, an playlist will be created from the
// user's favourited tracks
// Optionally, if the -s flag is set, a playlist will be created for each
// set (playlist) the user has created.
func fromUser(id uint64) {
	var filename string

	//Get users uploaded tracks
	tracks, err := api.User(id).Tracks(nil)
	if err != nil {
		panic(err)
	}

	//Get the user object so we can get their name
	user, err := api.User(id).Get(nil)
	if err != nil {
		panic(err)
	}

	//If the -f flag is set, also save the users favourited tracks to
	//a playlist
	if favouritesb {
		favs, err := api.User(id).Favorites(nil)
		if err != nil {
			panic(err)
		}

		filename = fmt.Sprintf("(Favourites) %v.m3u", user.Username)
		writeTracksToFile(filename, soundcloudTracksToGenericTracks(favs))
	}

	//If the -s flag is set, also save the users sets to m3u playlists
	if playlistsb {
		pls, err := api.User(id).Playlists(nil)
		if err != nil {
			panic(err)
		}

		for _, playlist := range pls {
			fromPlaylist(playlist.Id)
		}
	}

	filename = fmt.Sprintf("(Uploads) %v", user.Username)
	writeTracksToFile(filename, soundcloudTracksToGenericTracks(tracks))
}

// Save m3u from a playlist url
func fromPlaylist(id uint64) {
	pl, err := api.Playlist(id).Get(nil)
	if err != nil {
		handleError(err)
		return
	}

	filename := fmt.Sprintf("(Playlist) %v by %v", pl.Title, pl.SubUser.User.Username)
	writeTracksToFile(filename, soundcloudTracksToGenericTracks(pl.Tracks))
}

// Create file with filename, and write provided tracks to m3u playlist
func writeTracksToFile(filename string, tracks []Track) {
	//Make sure the tracks aren't empty before writing (maybe move
	//this somewhere else?)
	if len(tracks) == 0 {
		verboseMessage("Empty playlist, skipping: " + filename)
		return
	}

	//Open a file
	fp := filepath.Join(PlaylistDir, sanitize.Name(filename)+".m3u")
	f, err := os.Create(fp)
	if err != nil {
		handleError(err)
		return
	}
	defer f.Close()

	//Create writer for file
	w := bufio.NewWriter(f)

	//Write the extended m3u header
	fmt.Fprintln(w, "#EXTM3U")

	//Output the extended m3u data, followed by the stream url
	//Stream URL must be appended with a ClientID or else it will be
	//rejected
	for _, track := range tracks {
		fmt.Fprintf(w, "#EXTINF:%v,%v - %v\n", track.Duration, track.Title, track.Artist)
		fmt.Fprintf(w, "%v\n", track.Stream)
	}
	w.Flush()

	verboseMessage("Playlist saved: " + fp)
}

//soundcloudTracksToGenericTracks converts []soundcloud.Track to []Track.
//This is to have a common interface for any sort of streaming site we might
//pull from
func soundcloudTracksToGenericTracks(sctracks []*soundcloud.Track) []Track {
	var tracks []Track

	for _, sctrack := range sctracks {
		stream := sctrack.StreamUrl + "?client_id=" + ClientID
		artist := sctrack.SubUser.User.Username
		track := Track{Stream: stream, Artist: artist, Duration: int(sctrack.Duration / 1000), Title: sctrack.Title}
		tracks = append(tracks, track)
	}

	return tracks
}

// handleError to print meaningful error messages to the users
func handleError(err error) {
	switch err.Error() {
	case "empty location":
		fmt.Println("Error: Invalid URL provided (no resource found).")
	case "unknown resource":
		fmt.Println("Error: Unknown soundcloud resource received.")
		fmt.Println("Try using a track, profile, or playlist URL.")
	case "no url":
		fmt.Println("No URL provided. Use -u or -url option.")
	default:
		fmt.Println("Error: ", err)
	}
}

// Prints given messages only if the -verbose flag is set
func verboseMessage(message string) {
	if verboseb {
		fmt.Println(message)
	}
}

// Resolve parses the given Soundcloud URL, calls resolve() from the
// Soundcloud API, and returns the string of the requested resource
// (ie. track/playlist/user) as well as the resources ID
func Resolve(url string) (resource string, id uint64, err error) {
	resolvedUrl, err := api.Resolve(url)
	if err != nil {
		return
	}

	splitPath := strings.Split(strings.Split(resolvedUrl.Path, ".")[0], "/")

	resource = splitPath[1]
	id, _ = strconv.ParseUint(splitPath[2], 10, 64)

	return
}
