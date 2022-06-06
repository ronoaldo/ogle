// Command youtube allows you to interact with the Youtube service using a
// command line interface.
//
// Check the updated help with "youtube --help":
//
//	Usage of youtube:
//	-category category_id
//		The category_id of the video to update.
//	-channel channel_id
//		The channel_id id to use.
//	-cmd command
//		The command to execute. Use -cmd="list" to show all commands. (default "list")
//	-desc description
//		The description of the video to update.
//	-playlist playlist_id
//		The playlist_id to use.
//	-tags tags
//		The list of tags separated by ',' to be used in the updated video.
//	-title title
//		The title of the video to update.
//	-video video_id
//		The video_id to use for editing.
//
//	Use one of the following values for the -cmd parameter:
//		channels        list channels
//		subscribers     list subscribers
//		playlists       list playlists
//		playlist-items  list videos in playlist
//		playlist-dedup  remove duplicate videos from playlist
//		video-update    update details about a video
//		lives           list upcomming broadcasts
//		list,help       show this message
//		logout          revoke credentials
//
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/ronoaldo/ogle"
	"golang.org/x/net/context"
	"google.golang.org/api/youtube/v3"
)

// General command line options
var (
	command  string
	playlist string
	channel  string
	video    string
)

// Video update command line options
var (
	videoTitle       string
	videoDescription string
	videoCategory    string
	videoTags        string
)

// Globals
var (
	w   = ogle.NewTabWriter(os.Stdout)
	ctx = context.Background()
)

func init() {
	flag.StringVar(&command, "cmd", "list", "The `command` to execute. Use -cmd=\"list\" to show all commands.")
	flag.StringVar(&playlist, "playlist", "", "The `playlist_id` to use.")
	flag.StringVar(&channel, "channel", "", "The `channel_id` id to use.")
	flag.StringVar(&video, "video", "", "The `video_id` to use for editing.")
	flag.StringVar(&videoTitle, "title", "", "The `title` of the video to update.")
	flag.StringVar(&videoDescription, "desc", "", "The `description` of the video to update.")
	flag.StringVar(&videoCategory, "category", "", "The `category_id` of the video to update.")
	flag.StringVar(&videoTags, "tags", "", "The list of `tags` separated by ',' to be used in the updated video.")
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
	case "playlist-dedup", "dedup-playlist":
		removeDuplicatesFromPlaylist(yt)
	case "video-update":
		videoUpdate(yt)
	case "lives":
		listLives(yt)
	case "reauth", "logout":
		logout()
	case "list", "help":
		listCommands()
	default:
		log.Printf("Unknown command: '%s'", command)
		listCommands()
	}
}

var cmdList = `
Use one of the following values for the -cmd parameter:
	channels	list channels
	subscribers	list subscribers
	playlists	list playlists
	playlist-items	list videos in playlist
	playlist-dedup	remove duplicate videos from playlist
	video-update	update details about a video
	lives		list upcomming broadcasts
	list,help	show this message
	logout		revoke credentials
`

func listCommands() {
	flag.Usage()
	fmt.Fprintf(os.Stdout, cmdList)
}

func listChannels(yt *youtube.Service) {
	count := 0
	w.Println("#", "ID", "NAME", "LANGUAGE", "URL", "SUBSCRIBERS", "VIDEOS", "UPLOADS_PLAYLIST", "VIEWS")
	defer w.Flush()
	req := yt.Channels.List([]string{"id,snippet,statistics,contentDetails"}).Mine(true)
	err := req.Pages(ctx, func(resp *youtube.ChannelListResponse) error {
		for i := range resp.Items {
			ch := resp.Items[i]
			count++
			w.Println(count, ch.Id, ch.Snippet.Title, ch.Snippet.DefaultLanguage, ch.Snippet.CustomUrl,
				ch.Statistics.SubscriberCount, ch.Statistics.VideoCount,
				ch.ContentDetails.RelatedPlaylists.Uploads, ch.Statistics.ViewCount)
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

	w.Println("#", "PUBLISHED_AT", "VIDEO", "STATUS", "URL")
	defer w.Flush()

	req := yt.PlaylistItems.List([]string{"id,snippet,status,contentDetails"}).PlaylistId(playlist)

	err := req.Pages(context.Background(), func(resp *youtube.PlaylistItemListResponse) error {
		for _, item := range resp.Items {
			count++
			url := "https://youtu.be/" + item.ContentDetails.VideoId
			w.Println(count, item.Snippet.PublishedAt, item.Snippet.Title, item.Status.PrivacyStatus, url)
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

type byPubDate []*youtube.LiveBroadcast

func (b byPubDate) Len() int           { return len(b) }
func (b byPubDate) Less(i, j int) bool { return b[i].Snippet.PublishedAt < b[j].Snippet.PublishedAt }
func (b byPubDate) Swap(i, j int)      { b[j], b[i] = b[i], b[j] }

func listLives(yt *youtube.Service) {
	count := 0

	lives := make([]*youtube.LiveBroadcast, 0)
	req := yt.LiveBroadcasts.List([]string{"id,snippet,contentDetails,status"}).BroadcastStatus("all")
	err := req.Pages(ctx, func(resp *youtube.LiveBroadcastListResponse) error {
		for _, item := range resp.Items {
			lives = append(lives, item)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	sort.Sort(byPubDate(lives))

	w.Println("#", "PUBLISHED_AT", "ID", "TITLE", "STATUS", "URL")
	defer w.Flush()
	for _, item := range lives {
		count++
		url := "https://studio.youtube.com/video/" + item.Id + "/livestreaming"
		w.Println(count, item.Snippet.PublishedAt, item.Id, item.Snippet.Title,
			item.Status.LifeCycleStatus, url)
	}
}

func videoUpdate(yt *youtube.Service) {
	if video == "" {
		log.Fatal("No video_id provided. Use the -video flag to define what video we need to update.")
	}

	var videoPayload *youtube.Video
	parts := []string{"id,snippet"}
	req := yt.Videos.List(parts).Id(video)
	err := req.Pages(ctx, func(resp *youtube.VideoListResponse) error {
		if len(resp.Items) == 0 {
			return fmt.Errorf("No vídeos matched the provided id '%s'", video)
		}
		videoPayload = resp.Items[0]
		return nil
	})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	msg := "Updated Fields: "
	if videoTitle != "" {
		videoPayload.Snippet.Title = videoTitle
		msg = msg + fmt.Sprintf("[title: %v]", videoTitle)
	}
	if videoDescription != "" {
		videoPayload.Snippet.Description = videoDescription
		msg = msg + fmt.Sprintf("[description: %v]", videoDescription)
	}
	if videoCategory != "" {
		videoPayload.Snippet.CategoryId = videoCategory
		msg = msg + fmt.Sprintf("[categoryId: %v]", videoCategory)
	}
	if videoTags != "" {
		cleanedTags := []string{}
		for _, tag := range strings.Split(videoTags, ",") {
			cleanedTags = append(cleanedTags, strings.TrimSpace(tag))
		}
		videoPayload.Snippet.Tags = cleanedTags
		msg = msg + fmt.Sprintf("[tags: %v]", strings.Join(cleanedTags, ","))
	}

	log.Printf("Updating video (id=%s). %s", video, msg)
	_, err = yt.Videos.Update(parts, videoPayload).Do()
	if err != nil {
		log.Fatalf("Error updating video: %v", err)
	}
	log.Println("Vídeo updated")
	// dumpjson(updatedVideo)
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

func dumpjson(v interface{}) {
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}
