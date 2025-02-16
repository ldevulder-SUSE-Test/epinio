// Package client connects to the Epinio API's endpoints
package client

import (
	"github.com/go-logr/logr"
)

// Client provides functionality for talking to an Epinio API
// server
type Client struct {
	log      logr.Logger
	URL      string
	WsURL    string // only stored here for the memo, the websocket client is not part of the epinioapi, yet.
	user     string
	password string
}

// New returns a new Epinio API client
func New(log logr.Logger, url string, wsURL string, user string, password string) *Client {
	return &Client{log: log, URL: url, WsURL: wsURL, user: user, password: password}
}
