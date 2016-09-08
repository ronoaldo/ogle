package ogle

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	ClientID     string
	ClientSecret string
)

func init() {
	flag.StringVar(&ClientID, "client-id", "inform-your-client-id", "The `CLIENT_ID` to be used.")
	flag.StringVar(&ClientSecret, "client-secret", "inform-your-client-secret", "The `CLIENT_SECRET` to be used.")
}

func NewClient(ctx context.Context, api string, scopes ...string) (c *http.Client, err error) {
	var config = &oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
	var token *oauth2.Token

	if token, err = TokenFromCache(api); err != nil {
		log.Printf("Unable to used cached token (%v)", err)
		if token, err = TokenFromWeb(ctx, config); err != nil {
			return nil, err
		}
		if err := SaveTokenToCache(api, token); err != nil {
			log.Printf("Unable to save token to cache %v", err)
		}
	}
	return config.Client(ctx, token), nil
}

func tokenCacheFileName(api string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Caches", api+".token")
	case "linux", "freebsd":
		return filepath.Join(os.Getenv("HOME"), ".cache", api+".token")
	}
	return "."
}

func SaveTokenToCache(api string, token *oauth2.Token) error {
	filename := tokenCacheFileName(api)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(token)
}

func TokenFromCache(api string) (*oauth2.Token, error) {
	filename := tokenCacheFileName(api)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	t := new(oauth2.Token)
	err = gob.NewDecoder(f).Decode(t)
	return t, err
}

func TokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"
	authURL := config.AuthCodeURL(randState)
	fmt.Printf("Navigate to this URL to authorize:\n\n%s\n\n", authURL)

	var code string
	fmt.Printf("Paste the authorization code here: ")
	fmt.Scanf("%s", &code)

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return token, nil
}
