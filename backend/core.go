package backend

import (
	"net/http"
	"path/filepath"
	"time"
)

// Result is the outcome of downloading a single post.
type Result struct {
	MediaID string   `json:"media_id"`
	Kind    string   `json:"kind"` // "video" | "image" | "skipped" | "error"
	Files   []string `json:"files"`
	Message string   `json:"message"`
}

// Options controls a download.
type Options struct {
	Root      string
	Watermark bool
	HD        bool
}

// rateLimitDelay matches the ~1.9s the original sleeps before each download.
const rateLimitDelay = 1900 * time.Millisecond

// Downloader orchestrates per-post downloads.
type Downloader struct {
	client *http.Client
}

func NewDownloader() *Downloader {
	return &Downloader{client: &http.Client{Timeout: 90 * time.Second}}
}

// DownloadURL parses a URL and downloads it per the given options.
func (d *Downloader) DownloadURL(url string, opts Options) Result {
	link, err := ParseURL(url)
	if err != nil {
		return Result{Kind: "error", Message: err.Error()}
	}
	time.Sleep(rateLimitDelay)
	if opts.HD {
		return d.downloadHD(link, opts)
	}
	return d.downloadSD(link, opts)
}

func (d *Downloader) downloadSD(link TikTokLink, opts Options) Result {
	// Early dedup before the API call when the URL already gives us the username.
	if link.Username != "" && !link.IsPhoto() {
		suffix := "_Save"
		if opts.Watermark {
			suffix = "_Watermark"
		}
		if NewStorage(opts.Root, link.Username).Has(link.MediaID + suffix) {
			return Result{MediaID: link.MediaID, Kind: "skipped", Message: "Already downloaded"}
		}
	}

	info, err := GetMedia(d.client, link.MediaID, opts.Watermark)
	if err != nil {
		return Result{MediaID: link.MediaID, Kind: "error", Message: err.Error()}
	}
	if info == nil {
		return Result{MediaID: link.MediaID, Kind: "error", Message: "Could not resolve media"}
	}

	username := info.Username
	if username == "" {
		username = link.Username
	}
	store := NewStorage(opts.Root, username)
	var files []string

	switch {
	case info.IsImagePost():
		for i, imgURL := range info.ImageURLs {
			key := fmtKey(link.MediaID, i+1)
			if store.Has(key) {
				continue
			}
			dest := filepath.Join(store.ImagesDir, fmtName(link.MediaID, i+1, "jpeg"))
			if err := DownloadFile(d.client, imgURL, dest, 5); err != nil {
				return Result{MediaID: link.MediaID, Kind: "error", Message: err.Error()}
			}
			store.Mark(key)
			files = append(files, dest)
		}
		return Result{MediaID: link.MediaID, Kind: "image", Files: files}
	case info.VideoURL != "":
		suffix := "_Save"
		if opts.Watermark {
			suffix = "_Watermark"
		}
		key := link.MediaID + suffix
		if store.Has(key) {
			return Result{MediaID: link.MediaID, Kind: "skipped", Message: "Already downloaded"}
		}
		dest := filepath.Join(store.VideosDir, link.MediaID+suffix+".mp4")
		if err := DownloadFile(d.client, info.VideoURL, dest, 5); err != nil {
			return Result{MediaID: link.MediaID, Kind: "error", Message: err.Error()}
		}
		store.Mark(key)
		return Result{MediaID: link.MediaID, Kind: "video", Files: []string{dest}}
	default:
		return Result{MediaID: link.MediaID, Kind: "error", Message: "No media URL found"}
	}
}

func (d *Downloader) downloadHD(link TikTokLink, opts Options) Result {
	if link.IsPhoto() {
		hd, err := GetHdImages(d.client, link.MediaID, link.Username)
		if err != nil {
			return Result{MediaID: link.MediaID, Kind: "error", Message: err.Error()}
		}
		username := hd.Username
		if username == "" {
			username = link.Username
		}
		store := NewStorage(opts.Root, username)
		var files []string
		for i, imgURL := range hd.ImageURLs {
			key := fmtName(link.MediaID, i+1, "jpg")
			if store.Has(key) {
				continue
			}
			dest := filepath.Join(store.ImagesDir, key)
			if err := DownloadFile(d.client, imgURL, dest, 5); err != nil {
				return Result{MediaID: link.MediaID, Kind: "error", Message: err.Error()}
			}
			store.Mark(key)
			files = append(files, dest)
		}
		return Result{MediaID: link.MediaID, Kind: "image", Files: files}
	}

	// Early dedup for HD video before the tikwm task.
	if link.Username != "" && NewStorage(opts.Root, link.Username).Has(link.MediaID+"_HD") {
		return Result{MediaID: link.MediaID, Kind: "skipped", Message: "Already downloaded"}
	}
	hd, err := GetHdVideo(d.client, link.MediaID, link.Username)
	if err != nil {
		return Result{MediaID: link.MediaID, Kind: "error", Message: err.Error()}
	}
	username := hd.Username
	if username == "" {
		username = link.Username
	}
	store := NewStorage(opts.Root, username)
	key := link.MediaID + "_HD"
	if store.Has(key) {
		return Result{MediaID: link.MediaID, Kind: "skipped", Message: "Already downloaded"}
	}
	if hd.VideoURL == "" {
		return Result{MediaID: link.MediaID, Kind: "error", Message: "No HD video URL"}
	}
	dest := filepath.Join(store.VideosDir, link.MediaID+"_HD.mp4")
	if err := DownloadFile(d.client, hd.VideoURL, dest, 5); err != nil {
		return Result{MediaID: link.MediaID, Kind: "error", Message: err.Error()}
	}
	store.Mark(key)
	return Result{MediaID: link.MediaID, Kind: "video", Files: []string{dest}}
}

func fmtKey(mediaID string, n int) string  { return mediaID + "_" + itoa(n) }
func fmtName(mediaID string, n int, ext string) string {
	return mediaID + "_" + itoa(n) + "." + ext
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
