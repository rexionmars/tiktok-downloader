# TikTok Downloader — Desktop

![Wails](https://img.shields.io/badge/Wails-v2-DF0000)
![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)

A cross-platform **desktop GUI** for the TikTok downloader, built on the same
stack as the *Twitter/X Media Batch Downloader*: **Wails v2** (Go backend +
native WebView) with a **React 19 + Vite 7 + Tailwind 4 + shadcn/ui** frontend.

The download logic is a Go port of our Python CLI
([tiktok-downloader-py](../tiktok-downloader-py)) — same APIs, same on-disk
layout, same dedup behaviour — so there is no Python runtime to ship: the whole
app is one small native binary with the frontend embedded.

## Features

- **Single URL** — download one post (video or photo), with/without watermark.
- **Batch** — paste many URLs; live per-item progress via Wails events.
- **Profile** — scrape a whole user profile with headless Chrome (chromedp),
  then download everything.
- **HD** — high-quality downloads via the tikwm.com API (2-step task for video).
- **History** + native **output-folder picker**, light/dark theme, custom
  frameless title bar.

## Stack (mirrors the Twitter-X app)

| Layer | Tech |
|---|---|
| Desktop shell | Wails v2 — native WebView, single binary, `go:embed` frontend |
| Backend | Go (`backend/` package) |
| Profile scraping | chromedp (Chrome DevTools Protocol) |
| Frontend | React 19, TypeScript, Vite 7 |
| Styling | Tailwind CSS 4 (no config file) + OKLCH design tokens |
| Components | shadcn/ui (new-york) over Radix UI, lucide icons, sonner toasts |
| Go ↔ JS | Wails auto-generated typed bridge (`window.go.main.App`) |

## Project layout

```
tiktok-downloader-desktop/
  main.go              # Wails app entry (window, embed, bindings)
  app.go               # App struct — methods exposed to the React UI
  backend/             # Go port of the download logic
    urls.go            #   URL parse + short-link resolution
    official.go        #   official TikTok feed API (SD)
    tikwm.go           #   tikwm.com HD (images + 2-step video task)
    download.go        #   file download + per-user storage + dedup index
    core.go            #   orchestration (parse → API → download)
    scraper.go         #   profile scraping via chromedp
  frontend/
    src/
      App.tsx          # layout: TitleBar + Sidebar + page switch
      components/       # TitleBar, Sidebar, OutputBar, *Panel, ui/ (shadcn)
      lib/, types/
    wailsjs/           # generated Go↔JS bridge (committed for type-checking)
```

## Prerequisites

- **Go 1.23+**
- **Node + pnpm**
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Google Chrome / Chromium** (only for the Profile feature — chromedp drives it)

## Run

```bash
cd tiktok-downloader-desktop
wails dev      # hot-reload: Vite watches the frontend, Go rebuilds on change
```

`wails dev` regenerates the `frontend/wailsjs` bridge from the Go bindings. The
committed shims there exist only so the frontend type-checks before the first run.

## Download

Prebuilt binaries for Windows, macOS and Linux are published on the
[Releases page](https://github.com/rexionmars/tiktok-downloader/releases).
They are produced automatically by GitHub Actions on every `v*` tag
(see [.github/workflows/build.yml](.github/workflows/build.yml)).

> **Linux** requires `webkit2gtk-4.1`:
> `sudo apt install libwebkit2gtk-4.1-0` (Debian/Ubuntu).

## Build a native app

```bash
wails build                       # current platform
wails build -platform darwin/universal
wails build -platform windows/amd64
wails build -platform linux/amd64
# → build/bin/TikTokDownloaderDesktop(.app/.exe/…)
```

## Build status (verified)

- ✅ Frontend: `pnpm run build` (tsc strict + Vite) passes.
- ✅ Backend: `go build ./...` and `go vet ./...` pass; full binary (~6 MB) builds
  with the frontend embedded.
- ✅ Download logic verified live against the real TikTok API (valid MP4 written,
  dedup index created).

## Credits

Download logic derived from
[Jettcodey/TikTok-Downloader](https://github.com/Jettcodey/TikTok-Downloader)
(C#/WinForms). The desktop architecture follows the Wails + React pattern of the
*Twitter/X Media Batch Downloader*.

## License

Released under the [MIT License](LICENSE).

## Disclaimer

Educational, unofficial tool. Not affiliated with TikTok/ByteDance. TikTok and
tikwm.com change frequently and may rate-limit or block requests. Only download
content you have the right to download, and comply with TikTok's Terms of Service.
