"""HD downloads via the third-party tikwm.com API.

Ports the two HD code paths from MainForm.cs:

* HD images use the "default" endpoint:   GET https://www.tikwm.com/api/?url=<id>&hd=1
  -> data.images[*] (list of image URLs), data.author.unique_id

* HD videos use the newer task endpoint (2-step):
    POST https://www.tikwm.com/api/video/task/submit   (form: url=<id>, web=1)
    GET  https://www.tikwm.com/api/video/task/result?task_id=<id>   (poll)
  Ready when data.status == 2 and data.detail.size > 0.
  -> data.detail.play_url, data.detail.author.unique_id
"""

from __future__ import annotations

import time
from dataclasses import dataclass, field

import requests

HD_IMAGE_ENDPOINT = "https://www.tikwm.com/api/"
HD_VIDEO_SUBMIT = "https://www.tikwm.com/api/video/task/submit"
HD_VIDEO_RESULT = "https://www.tikwm.com/api/video/task/result"

_HEADERS = {
    "User-Agent": (
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
        "(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
    )
}


@dataclass
class HdResult:
    media_id: str
    username: str
    video_url: str | None = None
    image_urls: list[str] = field(default_factory=list)


class HdError(RuntimeError):
    """Raised when tikwm.com returns an error or never becomes ready."""


def get_hd_images(
    media_id_or_url: str,
    *,
    session: requests.Session | None = None,
    fallback_username: str | None = None,
    timeout: int = 30,
) -> HdResult:
    """Fetch HD image URLs for a photo post (or HD video url as a side effect)."""
    sess = session or requests.Session()
    resp = sess.get(
        HD_IMAGE_ENDPOINT,
        params={"url": media_id_or_url, "hd": "1"},
        headers=_HEADERS,
        timeout=timeout,
    )
    resp.raise_for_status()
    body = resp.json()
    if body.get("code") != 0:
        raise HdError(f"tikwm.com error for {media_id_or_url}: {body.get('msg')}")

    data = body.get("data") or {}
    username = (data.get("author") or {}).get("unique_id") or fallback_username or ""
    images = [str(u) for u in (data.get("images") or [])]
    return HdResult(
        media_id=str(media_id_or_url),
        username=username,
        image_urls=images,
        video_url=data.get("play"),  # present for video posts on this endpoint
    )


def get_hd_video(
    media_id_or_url: str,
    *,
    session: requests.Session | None = None,
    fallback_username: str | None = None,
    max_retries: int = 15,
    delay: float = 0.5,
    timeout: int = 30,
) -> HdResult:
    """Submit and poll the tikwm.com task endpoint for a high-bitrate HD video."""
    sess = session or requests.Session()

    submit = sess.post(
        HD_VIDEO_SUBMIT,
        data={"url": media_id_or_url, "web": "1"},
        headers=_HEADERS,
        timeout=timeout,
    )
    submit.raise_for_status()
    submit_body = submit.json()
    task_id = (submit_body.get("data") or {}).get("task_id")
    if submit_body.get("code") != 0 or not task_id:
        raise HdError(f"Failed to submit HD video task for {media_id_or_url}")

    for _ in range(max_retries):
        result = sess.get(
            HD_VIDEO_RESULT,
            params={"task_id": task_id},
            headers=_HEADERS,
            timeout=timeout,
        )
        result.raise_for_status()
        body = result.json()
        data = body.get("data") or {}
        if body.get("code") == 0 and data:
            detail = data.get("detail") or {}
            try:
                size = int(detail.get("size") or 0)
            except (TypeError, ValueError):
                size = 0
            if data.get("status") == 2 and size > 0:
                username = (
                    (detail.get("author") or {}).get("unique_id")
                    or fallback_username
                    or ""
                )
                return HdResult(
                    media_id=str(media_id_or_url),
                    username=username,
                    video_url=detail.get("play_url"),
                )
        time.sleep(delay)

    raise HdError(
        f"HD video task for {media_id_or_url} not ready after {max_retries} attempts"
    )
