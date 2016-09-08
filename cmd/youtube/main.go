package main

import (
	"flag"
	"fmt"
	"github.com/ronoaldo/ogle"
	"golang.org/x/net/context"
	"google.golang.org/api/youtube/v3"
	"log"
)

var (
	command string
)

func init() {
	flag.StringVar(&command, "c", "subscribers",
		"The `COMMAND` to execute. Currently supported are subscribers.")
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
	case "subscribers":
		listSubscribers(yt)
	default:
		flag.Usage()
	}
}

func listSubscribers(yt *youtube.Service) {
	nextPageToken := ""
	itemsPerPage := 50
	page := 1
	for {
		resp, err := yt.Subscriptions.
			List("subscriberSnippet").
			MySubscribers(true).
			PageToken(nextPageToken).
			Order("alphabetical").
			MaxResults(int64(itemsPerPage)).
			Do()
		if err != nil {
			log.Fatal(err)
		}
		for i := range resp.Items {
			sub := resp.Items[i]
			if sub.SubscriberSnippet != nil {
				fmt.Printf("%s,%s\n", sub.SubscriberSnippet.ChannelId, sub.SubscriberSnippet.Title)
			}
		}
		nextPageToken = resp.NextPageToken
		if nextPageToken == "" {
			break
		}
		page++
	}
}
