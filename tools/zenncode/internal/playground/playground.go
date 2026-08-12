// Package playground shares Go programs on the Go Playground.
//
// The share endpoint is content addressed: posting the same bytes twice
// returns the same identifier, so re-sharing an unchanged snippet reuses the
// link an article already has instead of minting a new one. That only holds
// when the body is sent as text/plain -- with a form content type the server
// stores the percent-encoded body and the identifier changes.
package playground

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	shareEndpoint = "https://go.dev/_/share"
	// LinkPrefix is the canonical form of a share link. play.golang.org
	// serves the same identifiers.
	LinkPrefix = "https://go.dev/play/p/"

	// MaxSize is the playground's limit on a shared program.
	MaxSize = 64 * 1024
)

var idRe = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)

// Client shares programs. The zero value is usable.
type Client struct {
	HTTP *http.Client
	// Pause is waited between successful shares to stay friendly to the
	// service during a bulk run.
	Pause time.Duration
}

// ErrTooLarge reports a program the playground will not accept.
var ErrTooLarge = errors.New("program exceeds the playground's size limit")

// Share uploads src and returns its playground link.
func (c *Client) Share(ctx context.Context, src string) (string, error) {
	if len(src) > MaxSize {
		return "", ErrTooLarge
	}
	const attempts = 3
	var lastErr error
	for attempt := range attempts {
		if attempt > 0 {
			select {
			case <-time.After(time.Duration(attempt) * 2 * time.Second):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		id, err := c.share(ctx, src)
		if err == nil {
			if c.Pause > 0 {
				select {
				case <-time.After(c.Pause):
				case <-ctx.Done():
					return "", ctx.Err()
				}
			}
			return LinkPrefix + id, nil
		}
		lastErr = err
		if !isRetryable(err) {
			break
		}
	}
	return "", lastErr
}

func (c *Client) share(ctx context.Context, src string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, shareEndpoint, strings.NewReader(src))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", &statusError{code: resp.StatusCode, body: strings.TrimSpace(string(body))}
	}
	id := strings.TrimSpace(string(body))
	if !idRe.MatchString(id) {
		return "", fmt.Errorf("share returned an unexpected id %q", id)
	}
	return id, nil
}

type statusError struct {
	code int
	body string
}

func (e *statusError) Error() string {
	if e.body == "" {
		return fmt.Sprintf("share failed: %s", http.StatusText(e.code))
	}
	return fmt.Sprintf("share failed: %s: %s", http.StatusText(e.code), e.body)
}

func isRetryable(err error) bool {
	var se *statusError
	if errors.As(err, &se) {
		return se.code == http.StatusTooManyRequests || se.code >= 500
	}
	return true // network-level failures
}
