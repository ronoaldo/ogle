package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/ronoaldo/ogle"
	"golang.org/x/net/context"
	"google.golang.org/api/youtube/v3"
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
	count := 0
	w := ogle.NewTabWriter(os.Stdout)
	defer w.Flush()
	w.Println("", "NAME", "CHANNEL_ID", "DESCRIPTION")
	for {
		resp, err := yt.Subscriptions.
			List([]string{"subscriberSnippet"}).
			MySubscribers(true).
			PageToken(nextPageToken).
			Order("alphabetical").
			MaxResults(int64(itemsPerPage)).
			Do()
		if err != nil {
			log.Fatal(err)
		}
		for i := range resp.Items {
			sub := resp.Items[i].SubscriberSnippet
			count++
			if sub != nil {
				desc := substr(sub.Description, 0, 40)
				w.Println(count, sub.Title, sub.ChannelId, strings.Split(desc, "\n")[0])
			}
		}
		nextPageToken = resp.NextPageToken
		if nextPageToken == "" {
			break
		}
		page++
	}
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
