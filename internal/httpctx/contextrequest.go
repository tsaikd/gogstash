package httpctx

import (
	"context"
	"io"
	"net/http"
)

// ClientPost issues a POST to the specified URL.
//
// Caller should close resp.Body when done reading from it.
//
// If the provided body is an io.Closer, it is closed after the
// request.
//
// To set custom headers, use NewRequestWithContext and Client.Do.
func ClientPost(ctx context.Context, c *http.Client, url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// Post issues a POST to the specified URL using DefaultClient.
//
// Caller should close resp.Body when done reading from it.
//
// If the provided body is an io.Closer, it is closed after the
// request.
//
// To set custom headers, use NewRequestWithContext and Client.Do.
func Post(ctx context.Context, url, contentType string, body io.Reader) (resp *http.Response, err error) {
	return ClientPost(ctx, http.DefaultClient, url, contentType, body)
}

// ClientGet issues a GET to the specified URL.
//
// Caller should close resp.Body when done reading from it.
func ClientGet(ctx context.Context, c *http.Client, url string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Get issues a GET to the specified URL using DefaultClient.
//
// Caller should close resp.Body when done reading from it.
func Get(ctx context.Context, url string) (resp *http.Response, err error) {
	return ClientGet(ctx, http.DefaultClient, url)
}
