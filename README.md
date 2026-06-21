# TikTok Downloader (Python port)

![Python](https://img.shields.io/badge/python-3.9%2B-blue)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)
![License](https://img.shields.io/badge/license-MIT-green)

Cross-platform **CLI** reimplementation of Jettcodey's
[TikTok Downloader](https://github.com/Jettcodey/TikTok-Downloader) (originally a
Windows-only C#/WinForms app). Runs on macOS, Linux and Windows.

It downloads TikTok videos and image posts:

- single post by URL (with or without watermark),
- in **HD** via the third-party `tikwm.com` API,
- in bulk from a `.txt` list of links,
- for an entire **user profile** (scraped with Playwright).

## Install

```bash
git clone https://github.com/rexionmars/tiktok-downloader.git
cd tiktok-downloader
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt

# Only needed for the `user` command:
python -m playwright install chromium
```

Or install as a package (gives you a `ttd` command):

```bash
pip install -e .
```

## Usage

Run via the package (`python -m tiktok_downloader`) or, if installed, as `ttd`.

```bash
# Single video, no watermark
python -m tiktok_downloader url "https://www.tiktok.com/@user/video/123456789"

# Single video, keep watermark
python -m tiktok_downloader url "<url>" --watermark

# Single post in HD (tikwm.com)
python -m tiktok_downloader url "<url>" --hd

# Image post (auto-detected) + the author's avatar
python -m tiktok_downloader url "<photo-url>" --avatars

# Short link (vm./vt.tiktok.com) — resolved automatically
python -m tiktok_downloader url "https://vm.tiktok.com/a1b2c3d4"

# Bulk from a text file (one URL per line)
python -m tiktok_downloader file links.txt --hd

# Whole profile: scrape links with a browser, then download them all
python -m tiktok_downloader user someuser
python -m tiktok_downloader user someuser --headless        # no visible window
```

### Options

| Flag | Meaning |
|------|---------|
| `-o, --output DIR` | Download root (default `./TikTokDownloads`) |
| `--hd` | Use the tikwm.com HD path |
| `--watermark` | Keep the watermark (SD only; ignored with `--hd`) |
| `--avatars` | Also download the author's avatar(s) |
| `--headless` | (`user` only) run the browser without a window |
| `--browser-path` | (`user` only) use a custom browser executable |

## Output layout

Mirrors the original app:

```
TikTokDownloads/
  <username>/
    Videos/   <id>_Save.mp4 | <id>_Watermark.mp4 | <id>_HD.mp4
    Images/   <id>_<n>.jpeg | <id>_<n>.jpg
    Avatars/  <username>_<n>.jpeg | <username>_GIF_<n>.gif
    <username>_index.txt        # dedup index — already-downloaded keys
  <username>_combined_links.txt # written by the `user` command
```

Re-running skips anything already listed in `<username>_index.txt`.

## How it maps to the original

| Original (C#) | Here (Python) |
|---|---|
| `GetMedia()` → `api22-normal-c-alisg.tiktokv.com/aweme/v1/feed/` | [`official_api.py`](tiktok_downloader/official_api.py) |
| HD images/videos via `tikwm.com` (2-step task for video) | [`tikwm_hd.py`](tiktok_downloader/tikwm_hd.py) |
| URL/id regexes + `vm./vt.` redirect resolution | [`urls.py`](tiktok_downloader/urls.py) |
| Playwright profile scroll + link collection | [`scraper.py`](tiktok_downloader/scraper.py) |
| File writes, retries, `_index.txt` dedup | [`downloader.py`](tiktok_downloader/downloader.py) |
| WinForms button handlers / flow | [`core.py`](tiktok_downloader/core.py) + [`cli.py`](tiktok_downloader/cli.py) |

**Note:** the original issues the feed request as HTTP `OPTIONS`; this port uses
`GET`, which returns the same feed and is more reliable.

## Caveats

TikTok and `tikwm.com` change frequently and may rate-limit or block requests;
the original project itself is no longer actively maintained. Endpoints can
break at any time. Use responsibly and only for content you have the right to
download.

## Credits

Port of [Jettcodey/TikTok-Downloader](https://github.com/Jettcodey/TikTok-Downloader)
(C#/WinForms). All original download logic, endpoints, and behaviour are derived
from that project.

## License

Released under the [MIT License](LICENSE).

## Disclaimer

This is an educational, unofficial tool. It is not affiliated with, endorsed by,
or sponsored by TikTok or ByteDance. You are solely responsible for how you use
it and for complying with TikTok's Terms of Service, applicable copyright law,
and the rights of content creators. Only download content you are authorized to
download.
