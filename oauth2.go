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

// NewClient creates a new http.Client that will authorize calls with the token
// stored for the given API name.
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

// SaveTokenToCache saves the given oauth2.Token to the cache file. It returns
// an error if the cache file cannot be written.
func SaveTokenToCache(api string, token *oauth2.Token) error {
	filename := tokenCacheFileName(api)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(token)
}

// LoadTokenFromCache reads the cache file for the provided API and returns a
// parsed oauth2.Token pointer, or an error if either the file cannot be read or
// the token is invalid.
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

// RemoveTokenFromCache will remove the cache file for the provided API. Future
// requests using that api name will require re-authentication.
func RemoveTokenFromCache(api string) error {
	if err := os.Remove(tokenCacheFileName(api)); err != nil {
		return err
	}
	return nil
}
