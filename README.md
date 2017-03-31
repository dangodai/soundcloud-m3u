# soundcloud-m3u

soundcloud-m3u is a command line tool that generates .m3u playlists, where the entries point to Soundcloud or Bandcamp http streams. The idea is to let you stream your favourite Soundcloud artists and playlists from any client you want (as long as they play http streams).

# Usage

soundcloud-m3u can handle three kinds of Soundcloud URLs: tracks, users, playlists/sets.

Likewise, it can handle tracks, albums, and users for Bandcamp links.

#### Example

```sh
$ soundcloud-m3u -u https://soundcloud.com/addivt -d ~/Music
```

```sh
$ soundcloud-m3u -u https://liluglymane.bandcamp.com/album/trick-dice -d ~/Music
```

#### Bandcamp

Bandcamp support is experimental currently. Most albums should generate m3u files fine, but some have issues. If you run into a problem, open an issue with the album causing problems.

#### Options

| Option | Explanation |
| ------ | ------ |
| -u, -url (Required) | The Soundcloud URL to generate the .m3u files from |
| -d, -dir | The directory to save the .m3u files to. Default: working directory |
| -s, -sets | If this flag is set, .m3u playlists will be generated for for all sets (albums and playlists) on the given user's profile. This only applies when a Soundcloud profile URL is given |
| -f, -favourites | If this flag is set, a .m3u playlist will be created from the given user's liked/favourited tracks. This only applies when a Soundcloud profile URL is given |
| -v, -verbose | Verbose logging |
| -id | See the client_id section below |

#### client_id

In order to use the Soundcloud API (including using the http stream), one needs a client\_id. A key is provided in the source code, however if you run into issues with this program, you will probably need to know more details about the client\_id.

Any user can apply for their own key, however the process is slow. Additionally, Soundcloud artists can choose to block the API from accessing their tracks. To get around this, we can find the Soundcloud master key in their app.js file. This is the key from their web player, and has no limitations. This key does change from time to time, so be prepared to find it or look for an update to this repository. Of course, if you don't want to break to API ToS get your own key. Using the provided key is your own choice.

# Install

Arch Linux (AUR): [soundcloud-m3u](https://aur.archlinux.org/packages/soundcloud-m3u)
```sh
yaourt -S soundcloud-m3u
```

Ubuntu
```sh
sudo add-apt-repository ppa:twodopeshaggy/ppapackages
sudo apt-get update
sudo apt-get install soundcloud-m3u
```

Debian
```sh
sudo apt-add-repository 'deb http://shaggytwodope.github.io/repo ./'
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 7086E9CC7EC3233B
sudo apt-key update
sudo apt-get update
sudo apt-get install soundcloud-m3u
```

# Building

soundclound-m3u is written in Go. This means you must have the Go compiler, and have your [GOPATH](https://golang.org/doc/code.html#GOPATH) set. To build simply run

```sh
$ go get -u github.com/dangodai/soundcloud-m3u
$ cd $GOPATH/src/github.com/dangodai/soundcloud-m3u
$ go install
```

The _go install_ command installs the binary to $GOPATH/bin.

### Todo
 - Improve the scraping of data from Bandcamp (currently ugly regex)
 - Provide binaries
