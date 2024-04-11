package transmission

import (
	"context"
	"net/url"
	"strings"

	"github.com/hekmon/cunits/v2"
	"github.com/hekmon/transmissionrpc/v3"
)

const endpoint = "/transmission/rpc"

type (
	// User to authenticate with Transmission
	User struct {
		Username string
		Password string
	}
	// Client connects to transmission via HTTP
	Client struct {
		tbt *transmissionrpc.Client
	}
)

// New create new transmission torrent
func New(baseUrl string, user *User) (*Client, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	u.Path = strings.TrimRight(u.Path, "/") + endpoint
	if user != nil {
		u.User = url.UserPassword(user.Username, user.Password)
	}

	tbt, err := transmissionrpc.New(u, nil)
	return &Client{
		tbt: tbt,
	}, err
}

// GetTorrents get a list of torrents
func (c *Client) GetTorrents(ctx context.Context) ([]transmissionrpc.Torrent, error) {
	return c.tbt.TorrentGet(
		ctx,
		[]string{
			"id",
			"name",
			"hashString",
			"status",
			"addedDate",
			"leftUntilDone",
			"eta",
			"uploadRatio",
			"rateDownload",
			"rateUpload",
			"downloadDir",
			"isFinished",
			"percentDone",
			"seedRatioMode",
			"error",
			"errorString",
			"files",
			"fileStats",
			"peers",
			"trackers",
			"trackerStats",
		},
		nil,
	)

}

// GetSession gets the current session from transmission
func (c *Client) GetSession(ctx context.Context) (transmissionrpc.SessionArguments, error) {
	return c.tbt.SessionArgumentsGetAll(ctx)
}

// GetSessionStats gets stats on the current & cumulative session
func (c *Client) GetSessionStats(ctx context.Context) (transmissionrpc.SessionStats, error) {
	return c.tbt.SessionStats(ctx)
}

// FreeSpace receives the free space statistics
func (c *Client) FreeSpace(ctx context.Context, path string) (free, totat cunits.Bits, err error) {
	return c.tbt.FreeSpace(ctx, path)
}
