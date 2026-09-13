package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

// Notifier delivers one reminder. Implementations are the MVP plugs:
// stdout, file, and webhook stub. Push/SMS SaaS is out of scope.
type Notifier interface {
	Notify(ctx context.Context, r Reminder) error
}

// WriterNotifier writes one JSON event per line to W (stdout when W is nil).
type WriterNotifier struct {
	W io.Writer
}

// Notify encodes r as a JSON line.
func (n WriterNotifier) Notify(_ context.Context, r Reminder) error {
	w := n.W
	if w == nil {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	return enc.Encode(r.Event())
}

// FileNotifier appends JSON lines to Path.
type FileNotifier struct {
	Path string
}

// Notify appends r as a JSON line, creating Path if needed.
func (n FileNotifier) Notify(_ context.Context, r Reminder) error {
	if n.Path == "" {
		return fmt.Errorf("file notifier: path is required")
	}
	f, err := os.OpenFile(n.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(r.Event())
}

// WebhookNotifier POSTs JSON Event bodies to URL. MVP stub — no retries,
// no SaaS vendor. Tests inject Client (e.g. httptest).
type WebhookNotifier struct {
	URL    string
	Client *http.Client
}

// Notify POSTs r.Event() to URL with Content-Type application/json.
func (n WebhookNotifier) Notify(ctx context.Context, r Reminder) error {
	if n.URL == "" {
		return fmt.Errorf("webhook notifier: url is required")
	}
	body, err := json.Marshal(r.Event())
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := n.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("webhook: HTTP %s", resp.Status)
	}
	return nil
}

// RecordingNotifier is a fake Notifier that stores what was sent.
type RecordingNotifier struct {
	mu   sync.Mutex
	sent []Reminder
}

// Notify records r.
func (n *RecordingNotifier) Notify(_ context.Context, r Reminder) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.sent = append(n.sent, r)
	return nil
}

// Sent returns a copy of recorded reminders.
func (n *RecordingNotifier) Sent() []Reminder {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]Reminder, len(n.sent))
	copy(out, n.sent)
	return out
}

// NewNotifier builds a MVP plug by name: stdout, file, or webhook.
func NewNotifier(kind, filePath, webhookURL string) (Notifier, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "", "stdout":
		return WriterNotifier{W: os.Stdout}, nil
	case "file":
		if filePath == "" {
			return nil, fmt.Errorf("notify: file notifier requires a path")
		}
		return FileNotifier{Path: filePath}, nil
	case "webhook":
		if webhookURL == "" {
			return nil, fmt.Errorf("notify: webhook notifier requires a url")
		}
		return WebhookNotifier{URL: webhookURL}, nil
	default:
		return nil, fmt.Errorf("notify: unknown notifier %q (stdout, file, webhook)", kind)
	}
}
