// Command youtube allows you to interact with the Youtube service using a
// command line interface.
//
// Check the updated help with `youtube --help`:
//
//     Usage of youtube:
//       -channel channel_id
//             The channel_id id to use.
//       -cmd command
//             The command to execute. Use -cmd="list" to show all commands. (default "list")
//       -playlist playlist_id
//             The playlist_id to use.
//
//     Valid commands for -cmd are:
//
//     For channels:
//             channels                list channels
//             subscribers             list subscribers
//
//     For playlists:
//             playlists               list playlists
//             playlist-items          list videos in playlist
//             dedup-playlist          remove duplicate videos from playlist
//
//     Misc:
//             list                    show this message
//             logout                  revoke credentials
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ronoaldo/ogle"
	"golang.org/x/net/context"
	"google.golang.org/api/youtube/v3"
)

// Command line options
var (
	command  string
	playlist string
	channel  string
)

// Globals
var (
	w   = ogle.NewTabWriter(os.Stdout)
	ctx = context.Background()
)

func init() {
	flag.StringVar(&command, "cmd", "list",
		"The `command` to execute. Use -cmd=\"list\" to show all commands.")
	flag.StringVar(&playlist, "playlist", "",
		"The `playlist_id` to use.")
	flag.StringVar(&channel, "channel", "",
		"The `channel_id` id to use.")
}

func main() {
	flag.Parse()

	client, err := ogle.NewClient(ctx, "youtube", youtube.YoutubeScope)
	if err != nil {
		log.Fatal(err)
	}

	yt, err := youtube.New(client)
	if err != nil {
		log.Fatal(err)
	}

	switch command {
	case "channels":
		listChannels(yt)
	case "subscribers", "subs":
		listSubscribers(yt)
	case "playlists":
		listPlaylists(yt)
	case "playlist-videos", "playlist-items":
		listPlaylistVideos(yt)
	case "dedup-playlist":
		removeDuplicatesFromPlaylist(yt)
	case "reauth", "logout":
		logout()
	case "list":
		listCommands()
	default:
		flag.Usage()
	}
}

var cmdList = `
Valid commands for -cmd are:

For channels:
	channels		list channels
	subscribers		list subscribers

For playlists:
	playlists		list playlists
	playlist-items		list videos in playlist
	dedup-playlist		remove duplicate videos from playlist

Misc:
	list			show this message
	logout			revoke credentials
`

func listCommands() {
	flag.Usage()
	fmt.Fprintf(os.Stdout, cmdList)
}

func listChannels(yt *youtube.Service) {
	count := 0
	w.Println("#", "ID", "NAME", "LANGUAGE", "URL", "SUBSCRIBERS", "VIDEOS", "VIEWS")
	defer w.Flush()
	req := yt.Channels.List([]string{"id,snippet,statistics"}).Mine(true)
	err := req.Pages(ctx, func(resp *youtube.ChannelListResponse) error {
		for i := range resp.Items {
			ch := resp.Items[i]
			count++
			w.Println(count, ch.Id, ch.Snippet.Title, ch.Snippet.DefaultLanguage, ch.Snippet.CustomUrl,
				ch.Statistics.SubscriberCount, ch.Statistics.VideoCount, ch.Statistics.ViewCount)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func listSubscribers(yt *youtube.Service) {
	count := 0
	w.Println("#", "NAME", "DESCRIPTION", "URL")
	defer w.Flush()
	req := yt.Subscriptions.List([]string{"subscriberSnippet"}).MySubscribers(true).Order("alphabetical")
	err := req.Pages(context.Background(), func(resp *youtube.SubscriptionListResponse) error {
		for _, sub := range resp.Items {
			count++
			desc := substr(sub.SubscriberSnippet.Description, 0, 40)
			url := "https://www.youtube.com/channel/" + sub.SubscriberSnippet.ChannelId
			w.Println(count, sub.SubscriberSnippet.Title, strings.Split(desc, "\n")[0], url)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func listPlaylists(yt *youtube.Service) {
	count := 0
	w.Println("#", "ID", "NAME", "CHANNEL", "VISIBILITY", "VIDEOS")
	defer w.Flush()

	req := yt.Playlists.List([]string{"id,snippet,status,contentDetails"})
	if channel != "" {
		req.ChannelId(channel)
	} else {
		req.Mine(true)
	}

	err := req.Pages(context.Background(), func(resp *youtube.PlaylistListResponse) error {
		for _, p := range resp.Items {
			count++
			w.Println(count, p.Id, p.Snippet.Title, p.Snippet.ChannelTitle,
				p.Status.PrivacyStatus, p.ContentDetails.ItemCount)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func listPlaylistVideos(yt *youtube.Service) {
	if playlist == "" {
		log.Fatal("You must spefify a playlist with `-playlist` argument.")
	}
	count := 0

	w.Println("#", "VIDEO", "STATUS", "URL")
	defer w.Flush()

	req := yt.PlaylistItems.List([]string{"id,snippet,status,contentDetails"}).PlaylistId(playlist)

	err := req.Pages(context.Background(), func(resp *youtube.PlaylistItemListResponse) error {
		for _, item := range resp.Items {
			count++
			url := "https://youtu.be/" + item.ContentDetails.VideoId
			w.Println(count, item.Snippet.Title, item.Status.PrivacyStatus, url)
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func removeDuplicatesFromPlaylist(yt *youtube.Service) {
	type strTuple [2]string
	itemID, videoID := 0, 1

	if playlist == "" {
		log.Fatal("You must specify a playlist with `-playlist` argument.")
	}
	req := yt.PlaylistItems.List([]string{"id,contentDetails"}).PlaylistId(playlist)

	videos := make([]strTuple, 0)
	toRemove := make([]strTuple, 0, len(videos))
	uniqueVids := make(map[string]strTuple, len(videos))

	err := req.Pages(ctx, func(resp *youtube.PlaylistItemListResponse) error {
		for _, item := range resp.Items {
			v := strTuple{item.Id, item.ContentDetails.VideoId}
			videos = append(videos, v)
			if _, isDup := uniqueVids[v[videoID]]; isDup {
				log.Printf("Duplicate video found with videoId=%v; itemId=%v", v[videoID], v[itemID])
				toRemove = append(toRemove, v)
				continue
			}
			uniqueVids[v[videoID]] = v
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(uniqueVids) != len(videos) {
		log.Printf("Removing duplicates from playlistId=%v, will keep %d videos (down from %d)",
			playlist, len(uniqueVids), len(videos))
		for _, v := range toRemove {
			id := v[itemID]
			log.Printf("> Will remove playlistItem %s", id)
			err := yt.PlaylistItems.Delete(id).Do()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("< Removed %v", id)
		}
		return
	}

	log.Println("Playlist has no duplicate videos", len(uniqueVids), len(videos))
}

func logout() {
	if err := ogle.RemoveTokenFromCache("youtube"); err != nil {
		log.Fatalf("Unable to remove authentication token: %v", err)
	}
	log.Println("Authentication token removed.")
}

// From: https://go.dev/play/p/SWY4Lu5Ano5
func substr(s string, from, length int) string {
	//create array like string view
	wb := []string{}
	wb = strings.Split(s, "")

	//miss nil pointer error
	to := from + length

	if to > len(wb) {
		to = len(wb)
	}

	if from > len(wb) {
		from = len(wb)
	}

	out := strings.Join(wb[from:to], "")
	if s == out {
		return s
	}
	return out + "..."
}
