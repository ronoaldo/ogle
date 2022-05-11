// package ogle provides some helper functions to work with Google APIs
package ogle

import (
	_ "embed"
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

//go:embed ogle.json
var oauthConfigData []byte

func newOAuth2Config(api string, scopes ...string) (*oauth2.Config, error) {
	return google.ConfigFromJSON(oauthConfigData, scopes...)
}

func NewClient(ctx context.Context, api string, scopes ...string) (c *http.Client, err error) {
	config, err := newOAuth2Config(api, scopes...)
	if err != nil {
		return nil, err
	}

	var token *oauth2.Token
	if token, err = LoadTokenFromCache(api); err != nil {
		log.Printf("Unable to reuse cached token: %v", err)
		if token, err = Authorize(ctx, config); err != nil {
			return nil, err
		}
		if err := SaveTokenToCache(api, token); err != nil {
			log.Printf("Unable to save token to cache: %v", err)
		}
	}

	return config.Client(ctx, token), nil
}

func tokenCacheFileName(api string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Caches", api+".token")
	case "linux", "freebsd":
		return filepath.Join(os.Getenv("HOME"), ".cache", "ogle-"+api+".token")
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

func LoadTokenFromCache(api string) (*oauth2.Token, error) {
	filename := tokenCacheFileName(api)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	t := new(oauth2.Token)
	err = gob.NewDecoder(f).Decode(t)
	return t, err
}
