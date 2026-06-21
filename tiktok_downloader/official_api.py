"""Official TikTok mobile feed API (SD / no-watermark / watermark).

Ports GetMedia() from MainForm.cs. The C# app issues an HTTP OPTIONS request to
the feed endpoint; in practice a GET returns the same JSON feed and is far more
reliable, so we use GET here.

JSON shape consumed (aweme_list[0]):
    video.play_addr.url_list[0]      -> no-watermark video
    video.download_addr.url_list[0]  -> watermarked video
    image_post_info.images[*].display_image.url_list[0] -> image post
    author.avatar_medium.url_list    -> avatar(s)
    author.video_icon.url_list       -> animated avatar(s)
    author.unique_id / unique_Id     -> username
"""

from __future__ import annotations

import time
from dataclasses import dataclass, field

import requests

API_URL = "https://api22-normal-c-alisg.tiktokv.com/aweme/v1/feed/"

# Device params copied verbatim from the C# implementation.
_API_PARAMS = {
    "iid": "7238789370386695942",
    "device_id": "7238787983025079814",
    "resolution": "1080*2400",
    "channel": "googleplay",
    "app_name": "musical_ly",
    "version_code": "350103",
    "device_platform": "android",
    "device_type": "Pixel 7",
    "os_version": "13",
}

_HEADERS = {
    "User-Agent": (
        "com.zhiliaoapp.musically/2023501030 (Linux; U; Android 13; en; "
        "Pixel 7; Build/TQ2A.230505.002; Cronet/58.0.2991.0)"
    )
}


@dataclass
class MediaInfo:
    """Resolved media for a single post (official API)."""

    media_id: str
    username: str
    video_url: str | None = None  # selected per watermark preference
    image_urls: list[str] = field(default_factory=list)
    avatar_urls: list[str] = field(default_factory=list)
    gif_avatar_urls: list[str] = field(default_factory=list)

    @property
    def is_image_post(self) -> bool:
        return not self.video_url and bool(self.image_urls)


def _first(url_list) -> str | None:
    if isinstance(url_list, list) and url_list:
        return url_list[0]
    return None


def get_media(
    media_id: str,
    *,
    watermark: bool = False,
    session: requests.Session | None = None,
    max_429_retries: int = 3,
    timeout: int = 30,
) -> MediaInfo | None:
    """Fetch media metadata for a single post from the official feed API.

    Returns None if the post can't be resolved (mismatch / empty / not found).
    Retries on HTTP 429 with a 5s delay, like the C# app.
    """
    sess = session or requests.Session()
    params = {"aweme_id": media_id, **_API_PARAMS}

    for attempt in range(max_429_retries + 1):
        resp = sess.get(API_URL, params=params, headers=_HEADERS, timeout=timeout)
        if resp.status_code == 429:
            if attempt < max_429_retries:
                time.sleep(5)
                continue
            return None
        resp.raise_for_status()

        if not resp.text.strip():
            return None
        data = resp.json()
        aweme_list = data.get("aweme_list") or []
        if not aweme_list:
            return None

        aweme = aweme_list[0]
        if str(aweme.get("aweme_id")) != str(media_id):
            return None

        author = aweme.get("author") or {}
        username = author.get("unique_id") or author.get("unique_Id") or ""

        video = aweme.get("video") or {}
        if watermark:
            video_url = _first((video.get("download_addr") or {}).get("url_list"))
        else:
            video_url = _first((video.get("play_addr") or {}).get("url_list"))

        image_urls: list[str] = []
        image_post = aweme.get("image_post_info") or {}
        for img in image_post.get("images") or []:
            u = _first((img.get("display_image") or {}).get("url_list"))
            if u:
                image_urls.append(u)

        return MediaInfo(
            media_id=media_id,
            username=username,
            video_url=video_url,
            image_urls=image_urls,
            avatar_urls=(author.get("avatar_medium") or {}).get("url_list") or [],
            gif_avatar_urls=(author.get("video_icon") or {}).get("url_list") or [],
        )

    return None
