package lnbot

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// EventsService handles the wallet event stream.
type EventsService struct{ c *Client }

// Stream opens an SSE stream of all wallet events.
// Events include invoice.created, invoice.settled, payment.created,
// payment.settled, and payment.failed. The channel is closed when the
// stream ends. Cancel the context to abort.
func (s *EventsService) Stream(ctx context.Context) (<-chan WalletEvent, <-chan error) {
	events := make(chan WalletEvent, 4)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		path := s.c.baseURL + "/v1/events"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err != nil {
			errs <- err
			return
		}
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("User-Agent", "lnbot-go/"+Version)
		if s.c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+s.c.apiKey)
		}

		resp, err := s.c.http.Do(req)
		if err != nil {
			errs <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			errs <- parseAPIError(resp.StatusCode, body)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			raw := strings.TrimSpace(line[5:])
			if raw == "" {
				continue
			}
			var ev WalletEvent
			if json.Unmarshal([]byte(raw), &ev) == nil {
				events <- ev
			}
		}

		if err := scanner.Err(); err != nil {
			errs <- err
		}
	}()

	return events, errs
}
