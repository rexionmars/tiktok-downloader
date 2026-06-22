package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tiktokdownloaderdesktop/backend"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound application struct. Its exported methods form the API
// surface the React frontend calls via the generated window.go.main.App bridge.
type App struct {
	ctx        context.Context
	dl         *backend.Downloader
	cancelMu   sync.Mutex
	cancelFunc context.CancelFunc
}

func NewApp() *App {
	return &App{dl: backend.NewDownloader()}
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }
func (a *App) shutdown(ctx context.Context) {
	a.cancelMu.Lock()
	if a.cancelFunc != nil {
		a.cancelFunc()
	}
	a.cancelMu.Unlock()
}

// ---- request/response types (mirror Twitter-X app.go style) -----------------

type DownloadRequest struct {
	URL       string `json:"url"`
	Output    string `json:"output"`
	HD        bool   `json:"hd"`
	Watermark bool   `json:"watermark"`
}

type BatchRequest struct {
	URLs      []string `json:"urls"`
	Output    string   `json:"output"`
	HD        bool     `json:"hd"`
	Watermark bool     `json:"watermark"`
}

type ProfileRequest struct {
	Username  string `json:"username"`
	Output    string `json:"output"`
	HD        bool   `json:"hd"`
	Watermark bool   `json:"watermark"`
}

// BatchProgress is emitted on the "batch:progress" event during batch/profile runs.
type BatchProgress struct {
	Index   int            `json:"index"`
	Total   int            `json:"total"`
	URL     string         `json:"url"`
	Result  backend.Result `json:"result"`
}

// ---- single URL -------------------------------------------------------------

// DownloadURL downloads one post and returns its result.
func (a *App) DownloadURL(req DownloadRequest) backend.Result {
	res := a.dl.DownloadURL(req.URL, backend.Options{
		Root:      a.resolveRoot(req.Output),
		HD:        req.HD,
		Watermark: req.Watermark,
	})
	a.recordHistory(req.URL, res)
	return res
}

// ---- batch ------------------------------------------------------------------

// DownloadBatch downloads a list of URLs sequentially, emitting "batch:progress"
// events the UI listens to. Returns all results.
func (a *App) DownloadBatch(req BatchRequest) []backend.Result {
	ctx, cancel := context.WithCancel(a.ctx)
	a.setCancel(cancel)
	defer a.clearCancel()

	root := a.resolveRoot(req.Output)
	results := make([]backend.Result, 0, len(req.URLs))
	for i, url := range req.URLs {
		select {
		case <-ctx.Done():
			return results
		default:
		}
		res := a.dl.DownloadURL(url, backend.Options{Root: root, HD: req.HD, Watermark: req.Watermark})
		results = append(results, res)
		a.recordHistory(url, res)
		runtime.EventsEmit(a.ctx, "batch:progress", BatchProgress{
			Index: i + 1, Total: len(req.URLs), URL: url, Result: res,
		})
	}
	return results
}

// StopBatch cancels an in-flight batch/profile download.
func (a *App) StopBatch() {
	a.cancelMu.Lock()
	if a.cancelFunc != nil {
		a.cancelFunc()
	}
	a.cancelMu.Unlock()
}

// ---- profile (scrape then download) ----------------------------------------

// ScrapeProfile scrapes a profile and returns the collected posts (URL +
// thumbnail). Always returns a non-nil slice so the JSON payload is [] (not
// null) on the JS side.
func (a *App) ScrapeProfile(username string) ([]backend.PostPreview, error) {
	ctx, cancel := context.WithCancel(a.ctx)
	a.setCancel(cancel)
	defer a.clearCancel()
	previews, err := backend.ScrapeUserPreviews(ctx, username, 200)
	if previews == nil {
		previews = []backend.PostPreview{}
	}
	return previews, err
}

// DownloadProfile scrapes a profile then downloads every collected post.
func (a *App) DownloadProfile(req ProfileRequest) ([]backend.Result, error) {
	previews, err := a.ScrapeProfile(req.Username)
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0, len(previews))
	for _, p := range previews {
		urls = append(urls, p.URL)
	}
	return a.DownloadBatch(BatchRequest{
		URLs: urls, Output: req.Output, HD: req.HD, Watermark: req.Watermark,
	}), nil
}

// DownloadURLs downloads a specific list of post URLs (e.g. a selection from a
// scraped profile grid).
func (a *App) DownloadURLs(req BatchRequest) []backend.Result {
	return a.DownloadBatch(req)
}

// ---- folders / system -------------------------------------------------------

// GetDefaultOutput returns the default download root (~/Downloads/TikTokDownloads).
func (a *App) GetDefaultOutput() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Downloads", "TikTokDownloads")
}

// SelectOutputFolder opens a native folder picker and returns the chosen path.
func (a *App) SelectOutputFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Select download folder",
		DefaultDirectory:     a.GetDefaultOutput(),
		CanCreateDirectories: true,
	})
}

// OpenFolder reveals a path in the system file manager.
func (a *App) OpenFolder(path string) {
	runtime.BrowserOpenURL(a.ctx, "file://"+path)
}

func (a *App) resolveRoot(output string) string {
	if output == "" {
		return a.GetDefaultOutput()
	}
	return output
}

// ---- history (simple JSON store) -------------------------------------------

type HistoryItem struct {
	URL       string `json:"url"`
	MediaID   string `json:"media_id"`
	Kind      string `json:"kind"`
	Timestamp int64  `json:"timestamp"`
}

func (a *App) historyPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".tiktok-downloader-desktop")
	os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "history.json")
}

// GetHistory returns the recent download history (most recent first).
func (a *App) GetHistory() []HistoryItem {
	data, err := os.ReadFile(a.historyPath())
	if err != nil {
		return []HistoryItem{}
	}
	var items []HistoryItem
	json.Unmarshal(data, &items)
	return items
}

// ClearHistory empties the history store.
func (a *App) ClearHistory() error { return os.Remove(a.historyPath()) }

func (a *App) recordHistory(url string, res backend.Result) {
	if res.Kind != "video" && res.Kind != "image" {
		return
	}
	items := a.GetHistory()
	item := HistoryItem{URL: url, MediaID: res.MediaID, Kind: res.Kind, Timestamp: time.Now().Unix()}
	items = append([]HistoryItem{item}, items...)
	if len(items) > 50 {
		items = items[:50]
	}
	if data, err := json.Marshal(items); err == nil {
		os.WriteFile(a.historyPath(), data, 0o644)
	}
}

func (a *App) setCancel(c context.CancelFunc) {
	a.cancelMu.Lock()
	a.cancelFunc = c
	a.cancelMu.Unlock()
}
func (a *App) clearCancel() {
	a.cancelMu.Lock()
	a.cancelFunc = nil
	a.cancelMu.Unlock()
}
