"""Command-line interface for the TikTok Downloader (Python port).

Subcommands:
    url       download a single post by URL
    file      download every URL listed in a .txt file
    user      scrape a user's profile (Playwright) then download all posts

Common flags:
    -o/--output   download root (default: ./TikTokDownloads)
    --hd          use the tikwm.com HD path
    --watermark   keep the watermark (SD only; ignored with --hd)
    --avatars     also download the author's avatar(s)
"""

from __future__ import annotations

import argparse
import sys

from .core import Downloader

DEFAULT_ROOT = "TikTokDownloads"


def _make_downloader(args) -> Downloader:
    return Downloader(
        root=args.output,
        watermark=getattr(args, "watermark", False),
        hd=getattr(args, "hd", False),
        verbose=True,
    )


def _summarize(results) -> int:
    ok = sum(1 for r in results if r.kind in ("video", "image"))
    skipped = sum(1 for r in results if r.kind == "skipped")
    errors = [r for r in results if r.kind == "error"]
    print(f"\nDone: {ok} downloaded, {skipped} skipped, {len(errors)} failed.")
    for r in errors:
        print(f"  error: {r.media_id or '?'}: {r.message}")
    return 1 if errors and ok == 0 else 0


def cmd_url(args) -> int:
    dl = _make_downloader(args)
    res = dl.download_url(args.url, with_avatars=args.avatars)
    return _summarize([res])


def cmd_file(args) -> int:
    dl = _make_downloader(args)
    results = dl.download_from_file(args.txt, with_avatars=args.avatars)
    return _summarize(results)


def cmd_user(args) -> int:
    from .scraper import scrape_user_links

    print(f"Scraping profile @{args.username} (a browser window may open)...")
    links_file = scrape_user_links(
        args.username,
        args.output,
        headless=args.headless,
        executable_path=args.browser_path,
    )
    print(f"Collected links -> {links_file}")

    dl = _make_downloader(args)
    results = dl.download_from_file(str(links_file), with_avatars=args.avatars)
    return _summarize(results)


def _add_common(p: argparse.ArgumentParser) -> None:
    p.add_argument(
        "-o", "--output", default=DEFAULT_ROOT, help="Download root folder"
    )
    p.add_argument("--hd", action="store_true", help="HD download via tikwm.com")
    p.add_argument(
        "--watermark",
        action="store_true",
        help="Keep watermark (SD only; ignored with --hd)",
    )
    p.add_argument(
        "--avatars", action="store_true", help="Also download author avatar(s)"
    )


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="ttd",
        description="Download TikTok videos and image posts (Python port).",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    p_url = sub.add_parser("url", help="Download a single post by URL")
    p_url.add_argument("url", help="TikTok post URL (long or vm./vt. short link)")
    _add_common(p_url)
    p_url.set_defaults(func=cmd_url)

    p_file = sub.add_parser("file", help="Download all URLs from a .txt file")
    p_file.add_argument("txt", help="Path to a text file with one URL per line")
    _add_common(p_file)
    p_file.set_defaults(func=cmd_file)

    p_user = sub.add_parser("user", help="Scrape a profile then download all posts")
    p_user.add_argument("username", help="TikTok username (with or without @)")
    p_user.add_argument(
        "--headless", action="store_true", help="Run the browser headless"
    )
    p_user.add_argument(
        "--browser-path", default=None, help="Path to a custom browser executable"
    )
    _add_common(p_user)
    p_user.set_defaults(func=cmd_user)

    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    try:
        return args.func(args)
    except KeyboardInterrupt:
        print("\nInterrupted.", file=sys.stderr)
        return 130


if __name__ == "__main__":
    raise SystemExit(main())
