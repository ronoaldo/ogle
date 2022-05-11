package ogle

import (
	_ "embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// Authorized launches the OAuth2 flow, by listening to a random port on
// localhost. It returns the granted oauth2.Token if the flow completes
// sucessfully.  It will report any error during the process, including the user
// not authorizing the client at all or an error during the token exchange.
func Authorize(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// 1. Listen to random local port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("ogle: error listening to random port: %v", err)
	}
	url := fmt.Sprintf("http://localhost:%v/_/", l.Addr().(*net.TCPAddr).Port)
	defer l.Close()

	// 2. Redirect user to authorization URL
	config.RedirectURL = url
	authURL := config.AuthCodeURL("ogle-auth", oauth2.AccessTypeOffline)
	fmt.Printf("Navigate to this URL to authorize:\n\n%s\n\n", authURL)

	// 3. Wait for the authorization to complete
	tokenChan := make(chan string)
	go waitForCode(tokenChan, l)
	for {
		select {
		case code := <-tokenChan:
			token, err := config.Exchange(ctx, code)
			if err != nil {
				return nil, err
			}
			return token, nil
		case <-time.After(time.Second * 60):
			return nil, fmt.Errorf("ogle: timeout waiting for authorization")
		}
	}
}

var (
	//go:embed msg/invalidtoken.html
	htmlInvalidToken []byte

	//go:embed msg/authsuccess.html
	htmlSuccessToken []byte
)

func waitForCode(ch chan string, l net.Listener) {
	http.HandleFunc("/_/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		state := q.Get("state")
		code := q.Get("code")
		if state != "ogle-auth" {
			w.Write(htmlInvalidToken)
			log.Fatalf("Invalid state: %v", state)
			return
		}
		w.Write(htmlSuccessToken)
		ch <- code
	})
	http.Serve(l, nil)
}
