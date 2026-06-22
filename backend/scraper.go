package backend

import (
	"context"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// PostPreview is a scraped post link plus a best-effort thumbnail.
type PostPreview struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Kind      string `json:"kind"` // "video" | "photo"
}

// scrapeJS collects every post <a> on the page together with the thumbnail
// <img> found inside its card and whether it's a video or photo.
const scrapeJS = `
(() => {
  const out = [];
  document.querySelectorAll('a').forEach(a => {
    if (!a.href) return;
    const isVideo = a.href.includes('/video/');
    const isPhoto = a.href.includes('/photo/');
    if (!isVideo && !isPhoto) return;
    // Find a thumbnail: an <img> inside the link, or inside the nearest card.
    let img = a.querySelector('img');
    if (!img) {
      const card = a.closest('div');
      if (card) img = card.querySelector('img');
    }
    out.push({
      url: a.href,
      thumbnail: img ? (img.src || img.getAttribute('data-src') || '') : '',
      kind: isVideo ? 'video' : 'photo',
    });
  });
  return out;
})()
`

// ScrapeUserLinks returns just the post URLs (used by DownloadProfile).
func ScrapeUserLinks(ctx context.Context, username string, maxScrolls int) ([]string, error) {
	previews, err := ScrapeUserPreviews(ctx, username, maxScrolls)
	urls := make([]string, 0, len(previews))
	for _, p := range previews {
		urls = append(urls, p.URL)
	}
	return urls, err
}

// ScrapeUserPreviews opens a TikTok profile in Chrome (via chromedp), scrolls
// until the page stops yielding new posts, and returns the de-duplicated posts
// (URL + thumbnail) belonging to that user. Runs the browser *visible* because
// TikTok serves a near-empty page to headless browsers.
func ScrapeUserPreviews(ctx context.Context, username string, maxScrolls int) ([]PostPreview, error) {
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	baseURL := "https://www.tiktok.com/@" + username

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(desktopUA),
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.WindowSize(1280, 900),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	if err := chromedp.Run(browserCtx,
		chromedp.Navigate(baseURL),
		chromedp.Sleep(6*time.Second),
	); err != nil {
		return nil, err
	}

	// Accumulate across scrolls; TikTok virtualizes the grid so cards leave the
	// DOM as you scroll. Keep the first thumbnail we see for each URL.
	type entry struct {
		thumb string
		kind  string
	}
	seen := map[string]*entry{}
	var order []string

	collect := func() {
		var batch []PostPreview
		if err := chromedp.Run(browserCtx, chromedp.Evaluate(scrapeJS, &batch)); err != nil {
			return
		}
		for _, p := range batch {
			e, ok := seen[p.URL]
			if !ok {
				seen[p.URL] = &entry{thumb: p.Thumbnail, kind: p.Kind}
				order = append(order, p.URL)
			} else if e.thumb == "" && p.Thumbnail != "" {
				e.thumb = p.Thumbnail // backfill a missing thumbnail
			}
		}
	}

	collect()
	stale := 0
	for i := 0; i < maxScrolls; i++ {
		before := len(seen)
		if err := chromedp.Run(browserCtx,
			chromedp.Evaluate(`window.scrollBy(0, document.body.scrollHeight)`, nil),
			chromedp.Sleep(2500*time.Millisecond),
		); err != nil {
			break
		}
		collect()
		if len(seen) == before {
			stale++
			if stale >= 3 {
				break
			}
		} else {
			stale = 0
		}
	}

	// Keep only this user's posts, preserve discovery order.
	previews := []PostPreview{}
	for _, u := range order {
		if strings.Contains(u, "/@"+username+"/") || strings.Contains(u, "/"+username+"/") {
			e := seen[u]
			previews = append(previews, PostPreview{URL: u, Thumbnail: e.thumb, Kind: e.kind})
		}
	}
	return previews, nil
}
