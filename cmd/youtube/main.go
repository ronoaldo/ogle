// Command youtube allows you to interact with the Youtube service using a
// command line interface.
//
// See all the available options from the tool integrated help menu:
//
//     youtube --help
//
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

const (
	itemsPerPage int64 = 50
)

var (
	command  string
	playlist string
	channel  string

	w = ogle.NewTabWriter(os.Stdout)
)

func init() {
	flag.StringVar(&command, "cmd", "channels",
		"The `COMMAND` to execute. Use \"list\" to show all commands.")
	flag.StringVar(&playlist, "playlist", "",
		"The `PLAYLIST` to use.")
	flag.StringVar(&channel, "channel", "",
		"The `CHANNELID` id to use.")
}

func main() {
	flag.Parse()
	ctx := context.Background()

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
	case "subscribers":
		listSubscribers(yt)
	case "playlists":
		listPlaylists(yt)
	case "playlist-videos":
		listPlaylistVideos(yt)
	case "reauth":
		reauth()

	case "list":
		cmdList := `
Valid commands are:
	channels		list channels
	subscribers		list subscribers
	playlists		list playlists
	playlist-videos	list videos in playlist
	reauth			revoke credentials
`
		fmt.Fprintf(os.Stdout, cmdList)
	default:
		flag.Usage()
	}
}

func listChannels(yt *youtube.Service) {
	nextPageToken := ""
	count := 0
	w.Println("#", "ID", "NAME", "LANGUAGE", "URL", "SUBSCRIBERS", "VIDEOS", "VIEWS")
	defer w.Flush()
	for {
		resp, err := yt.Channels.List([]string{"id,snippet,statistics"}).Mine(true).
			PageToken(nextPageToken).MaxResults(itemsPerPage).Do()
		if err != nil {
			log.Fatal(err)
		}
		for i := range resp.Items {
			ch := resp.Items[i]
			count++
			w.Println(count, ch.Id, ch.Snippet.Title, ch.Snippet.DefaultLanguage, ch.Snippet.CustomUrl,
				ch.Statistics.SubscriberCount, ch.Statistics.VideoCount, ch.Statistics.ViewCount)
		}
		if nextPageToken = resp.NextPageToken; nextPageToken == "" {
			break
		}
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

func reauth() {
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
