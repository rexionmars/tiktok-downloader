"""TikTok URL parsing / normalization.

Ports the URL handling from MainForm.cs:
- video/photo id extraction via regex
- short-link (vm./vt.tiktok.com) resolution by following the redirect
"""

from __future__ import annotations

import re
from dataclasses import dataclass

import requests

# Same patterns the C# app uses (MainForm.cs).
_VIDEO_ID_RE = re.compile(r"/video/(\d+)")
_PHOTO_ID_RE = re.compile(r"/photo/(\d+)")
_USERNAME_RE = re.compile(r"/@([\w.]+)")

SHORT_LINK_HOSTS = ("vm.tiktok.com", "vt.tiktok.com")

# A desktop UA is enough to follow the short-link redirect.
_REDIRECT_HEADERS = {
    "User-Agent": (
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
        "(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
    )
}


@dataclass
class TikTokLink:
    """A resolved TikTok post link."""

    url: str  # the full (redirect-resolved) URL
    media_id: str  # the numeric aweme id
    kind: str  # "video" or "photo"
    username: str | None = None  # @handle without the leading @

    @property
    def is_photo(self) -> bool:
        return self.kind == "photo"


def is_short_link(url: str) -> bool:
    return any(host in url for host in SHORT_LINK_HOSTS)


def resolve_short_link(url: str, timeout: int = 30) -> str:
    """Follow vm./vt.tiktok.com redirects and return the final URL.

    Mirrors GetRedirectUrl() in the C# app: a GET that lets requests follow
    redirects, then reads the final resolved URL.
    """
    resp = requests.get(
        url, headers=_REDIRECT_HEADERS, allow_redirects=True, timeout=timeout
    )
    return resp.url


def extract_username(url: str) -> str | None:
    match = _USERNAME_RE.search(url)
    return match.group(1) if match else None


def parse(url: str, resolve: bool = True, timeout: int = 30) -> TikTokLink:
    """Parse a TikTok post URL into a TikTokLink.

    If `resolve` is True, short links are expanded first. Raises ValueError when
    no video/photo id can be found.
    """
    url = url.strip()
    if resolve and is_short_link(url):
        url = resolve_short_link(url, timeout=timeout)

    video_match = _VIDEO_ID_RE.search(url)
    if video_match:
        return TikTokLink(
            url=url,
            media_id=video_match.group(1),
            kind="video",
            username=extract_username(url),
        )

    photo_match = _PHOTO_ID_RE.search(url)
    if photo_match:
        return TikTokLink(
            url=url,
            media_id=photo_match.group(1),
            kind="photo",
            username=extract_username(url),
        )

    raise ValueError(f"Could not find a video/photo id in URL: {url!r}")
