package backend

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Storage manages per-user folders plus the dedup index file, mirroring the
// layout of the original C#/Python downloader.
type Storage struct {
	Username   string
	UserDir    string
	VideosDir  string
	ImagesDir  string
	AvatarsDir string
	indexPath  string
	index      map[string]bool
}

func safeUsername(username string) string {
	name := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(username), "@"))
	if name == "" {
		return "unknown"
	}
	return name
}

// NewStorage builds the storage paths for a user under root.
func NewStorage(root, username string) *Storage {
	u := safeUsername(username)
	userDir := filepath.Join(root, u)
	return &Storage{
		Username:   u,
		UserDir:    userDir,
		VideosDir:  filepath.Join(userDir, "Videos"),
		ImagesDir:  filepath.Join(userDir, "Images"),
		AvatarsDir: filepath.Join(userDir, "Avatars"),
		indexPath:  filepath.Join(userDir, u+"_index.txt"),
	}
}

func (s *Storage) loadIndex() {
	if s.index != nil {
		return
	}
	s.index = make(map[string]bool)
	f, err := os.Open(s.indexPath)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			s.index[line] = true
		}
	}
}

// Has reports whether a dedup key was already downloaded.
func (s *Storage) Has(key string) bool {
	s.loadIndex()
	return s.index[key]
}

// Mark records a dedup key in memory and appends it to the index file.
func (s *Storage) Mark(key string) error {
	s.loadIndex()
	s.index[key] = true
	if err := os.MkdirAll(s.UserDir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(s.indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, key)
	return err
}

// DownloadFile streams a URL to dest with retries, writing atomically via a
// temp file. Mirrors DownloadVideoWithBufferedWrite() in the original app.
func DownloadFile(client *http.Client, url, dest string, retries int) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	var lastErr error
	for attempt := 0; attempt < retries; attempt++ {
		err := func() error {
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("User-Agent", desktopUA)
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status %d", resp.StatusCode)
			}
			tmp := dest + ".part"
			out, err := os.Create(tmp)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, resp.Body); err != nil {
				out.Close()
				return err
			}
			out.Close()
			return os.Rename(tmp, dest)
		}()
		if err == nil {
			return nil
		}
		lastErr = err
		if attempt < retries-1 {
			time.Sleep(1500 * time.Millisecond)
		}
	}
	return fmt.Errorf("failed to download %s after %d attempts: %w", url, retries, lastErr)
}
