// Package backend ports the TikTok download logic from the Python CLI
// (tiktok-downloader-py) to Go, for use behind the Wails desktop UI.
package backend

import (
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	videoIDRe = regexp.MustCompile(`/video/(\d+)`)
	photoIDRe = regexp.MustCompile(`/photo/(\d+)`)
	usernRe   = regexp.MustCompile(`/@([\w.]+)`)
)

var shortLinkHosts = []string{"vm.tiktok.com", "vt.tiktok.com"}

const desktopUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// TikTokLink is a resolved TikTok post link.
type TikTokLink struct {
	URL      string `json:"url"`
	MediaID  string `json:"media_id"`
	Kind     string `json:"kind"` // "video" | "photo"
	Username string `json:"username"`
}

// IsPhoto reports whether the link points to a photo (slideshow) post.
func (l TikTokLink) IsPhoto() bool { return l.Kind == "photo" }

func isShortLink(url string) bool {
	for _, h := range shortLinkHosts {
		if strings.Contains(url, h) {
			return true
		}
	}
	return false
}

// resolveShortLink follows vm./vt.tiktok.com redirects and returns the final URL.
func resolveShortLink(url string, timeout time.Duration) (string, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", desktopUA)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return resp.Request.URL.String(), nil
}

func extractUsername(url string) string {
	if m := usernRe.FindStringSubmatch(url); m != nil {
		return m[1]
	}
	return ""
}

// ParseURL parses a TikTok post URL into a TikTokLink, resolving short links.
func ParseURL(url string) (TikTokLink, error) {
	url = strings.TrimSpace(url)
	if isShortLink(url) {
		if resolved, err := resolveShortLink(url, 30*time.Second); err == nil {
			url = resolved
		}
	}

	if m := videoIDRe.FindStringSubmatch(url); m != nil {
		return TikTokLink{URL: url, MediaID: m[1], Kind: "video", Username: extractUsername(url)}, nil
	}
	if m := photoIDRe.FindStringSubmatch(url); m != nil {
		return TikTokLink{URL: url, MediaID: m[1], Kind: "photo", Username: extractUsername(url)}, nil
	}
	return TikTokLink{}, &ParseError{URL: url}
}

// ParseError signals that no video/photo id could be found in a URL.
type ParseError struct{ URL string }

func (e *ParseError) Error() string {
	return "could not find a video/photo id in URL: " + e.URL
}
