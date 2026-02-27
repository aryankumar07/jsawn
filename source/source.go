package source

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Source represents a resolved JSON input with its label and data or error.
type Source struct {
	Label string
	Data  []byte
	Err   error
}

// ResolveAll splits a comma-separated arg into individual sources, resolves each,
// and prepends stdin data if present.
func ResolveAll(arg string, stdinData []byte) []Source {
	var sources []Source

	if len(stdinData) > 0 {
		sources = append(sources, Source{
			Label: "stdin",
			Data:  stdinData,
		})
	}

	if arg != "" {
		parts := strings.Split(arg, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			sources = append(sources, resolve(p))
		}
	}

	return sources
}

func resolve(s string) Source {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return fetchURL(s)
	}
	return readFile(s)
}

func fetchURL(rawURL string) Source {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return Source{
			Label: urlLabel(rawURL),
			Err:   fmt.Errorf("fetch failed: %w", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Source{
			Label: urlLabel(rawURL),
			Err:   fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status),
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Source{
			Label: urlLabel(rawURL),
			Err:   fmt.Errorf("failed to read response: %w", err),
		}
	}

	return Source{
		Label: urlLabel(rawURL),
		Data:  data,
	}
}

func readFile(path string) Source {
	data, err := os.ReadFile(path)
	if err != nil {
		return Source{
			Label: filepath.Base(path),
			Err:   fmt.Errorf("failed to read file: %w", err),
		}
	}
	return Source{
		Label: filepath.Base(path),
		Data:  data,
	}
}

func urlLabel(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	label := u.Host + u.Path
	// Trim trailing slash
	label = strings.TrimRight(label, "/")
	return label
}
