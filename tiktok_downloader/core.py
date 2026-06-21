"""High-level download orchestration.

Ties together urls -> official_api / tikwm_hd -> downloader, reproducing the
per-post flow of the C# MainForm (SD/HD, video/image, avatars, dedup, 1.9s
rate-limit delay between posts).
"""

from __future__ import annotations

import time
from dataclasses import dataclass

import requests

from . import official_api, tikwm_hd, urls
from .downloader import Storage, download_file

# The C# app sleeps ~1.9s before each download to dodge rate limits.
RATE_LIMIT_DELAY = 1.9


@dataclass
class DownloadResult:
    media_id: str
    kind: str  # "video" | "image" | "skipped" | "error"
    files: list[str]
    message: str = ""


class Downloader:
    def __init__(
        self,
        root: str,
        *,
        watermark: bool = False,
        hd: bool = False,
        rate_limit_delay: float = RATE_LIMIT_DELAY,
        verbose: bool = True,
    ):
        self.root = root
        self.watermark = watermark
        self.hd = hd
        self.rate_limit_delay = rate_limit_delay
        self.verbose = verbose
        self.session = requests.Session()

    def _log(self, msg: str) -> None:
        if self.verbose:
            print(msg)

    # --- public API -----------------------------------------------------
    def download_url(self, url: str, *, with_avatars: bool = False) -> DownloadResult:
        try:
            link = urls.parse(url)
        except ValueError as exc:
            return DownloadResult("", "error", [], str(exc))

        if self.rate_limit_delay:
            time.sleep(self.rate_limit_delay)

        if self.hd:
            return self._download_hd(link, with_avatars=with_avatars)
        return self._download_sd(link, with_avatars=with_avatars)

    # --- SD (official API) ---------------------------------------------
    def _download_sd(self, link: urls.TikTokLink, *, with_avatars: bool) -> DownloadResult:
        # Dedup before the API call when the URL already tells us the username,
        # so an already-downloaded video isn't re-fetched (matches the C# app,
        # which checks the index first). Only the single-video key is knowable
        # up front; image posts still need the API to count their images.
        if link.username and not link.is_photo:
            suffix = "_Watermark" if self.watermark else "_Save"
            pre = Storage(self.root, link.username)
            if pre.has(f"{link.media_id}{suffix}"):
                return DownloadResult(link.media_id, "skipped", [], "Already downloaded")

        info = official_api.get_media(
            link.media_id, watermark=self.watermark, session=self.session
        )
        if info is None:
            return DownloadResult(link.media_id, "error", [], "Could not resolve media")

        store = Storage(self.root, info.username or link.username)
        files: list[str] = []

        if info.is_image_post:
            for i, img_url in enumerate(info.image_urls, start=1):
                key = f"{link.media_id}_{i}"
                if store.has(key):
                    continue
                dest = store.images_dir / f"{link.media_id}_{i}.jpeg"
                self._log(f"Downloading image {dest.name}")
                download_file(img_url, dest, session=self.session)
                store.mark(key)
                files.append(str(dest))
            kind = "image"
        elif info.video_url:
            suffix = "_Watermark" if self.watermark else "_Save"
            key = f"{link.media_id}{suffix}"
            dest = store.videos_dir / f"{link.media_id}{suffix}.mp4"
            if store.has(key):
                return DownloadResult(link.media_id, "skipped", [], "Already downloaded")
            self._log(f"Downloading video {dest.name}")
            download_file(info.video_url, dest, session=self.session)
            store.mark(key)
            files.append(str(dest))
            kind = "video"
        else:
            return DownloadResult(link.media_id, "error", [], "No media URL found")

        if with_avatars:
            files += self._download_avatars(store, info)
        return DownloadResult(link.media_id, kind, files)

    # --- HD (tikwm.com) -------------------------------------------------
    def _download_hd(self, link: urls.TikTokLink, *, with_avatars: bool) -> DownloadResult:
        # Early dedup for HD video before the tikwm.com task, when the URL
        # already gives us the username.
        if link.username and not link.is_photo:
            pre = Storage(self.root, link.username)
            if pre.has(f"{link.media_id}_HD"):
                return DownloadResult(link.media_id, "skipped", [], "Already downloaded")
        try:
            if link.is_photo:
                hd = tikwm_hd.get_hd_images(
                    link.media_id,
                    session=self.session,
                    fallback_username=link.username,
                )
                store = Storage(self.root, hd.username or link.username)
                files: list[str] = []
                for i, img_url in enumerate(hd.image_urls, start=1):
                    key = f"{link.media_id}_{i}.jpg"
                    if store.has(key):
                        continue
                    dest = store.images_dir / f"{link.media_id}_{i}.jpg"
                    self._log(f"Downloading HD image {dest.name}")
                    download_file(img_url, dest, session=self.session)
                    store.mark(key)
                    files.append(str(dest))
                return DownloadResult(link.media_id, "image", files)
            else:
                hd = tikwm_hd.get_hd_video(
                    link.media_id,
                    session=self.session,
                    fallback_username=link.username,
                )
                store = Storage(self.root, hd.username or link.username)
                key = f"{link.media_id}_HD"
                dest = store.videos_dir / f"{link.media_id}_HD.mp4"
                if store.has(key):
                    return DownloadResult(
                        link.media_id, "skipped", [], "Already downloaded"
                    )
                if not hd.video_url:
                    return DownloadResult(link.media_id, "error", [], "No HD video URL")
                self._log(f"Downloading HD video {dest.name}")
                download_file(hd.video_url, dest, session=self.session)
                store.mark(key)
                return DownloadResult(link.media_id, "video", [str(dest)])
        except (tikwm_hd.HdError, requests.RequestException) as exc:
            return DownloadResult(link.media_id, "error", [], str(exc))

    # --- avatars --------------------------------------------------------
    def _download_avatars(self, store: Storage, info: official_api.MediaInfo) -> list[str]:
        files: list[str] = []
        for i, av_url in enumerate(info.avatar_urls, start=1):
            key = f"{store.username}_{i}"
            if store.has(key):
                continue
            dest = store.avatars_dir / f"{store.username}_{i}.jpeg"
            download_file(av_url, dest, session=self.session)
            store.mark(key)
            files.append(str(dest))
        for i, gif_url in enumerate(info.gif_avatar_urls, start=1):
            key = f"{store.username}_GIF_{i}"
            if store.has(key):
                continue
            dest = store.avatars_dir / f"{store.username}_GIF_{i}.gif"
            download_file(gif_url, dest, session=self.session)
            store.mark(key)
            files.append(str(dest))
        return files

    # --- batch helpers --------------------------------------------------
    def download_from_file(self, txt_path: str, *, with_avatars: bool = False):
        from pathlib import Path

        lines = [
            ln.strip()
            for ln in Path(txt_path).read_text(encoding="utf-8").splitlines()
            if ln.strip()
        ]
        results = []
        for i, url in enumerate(lines, start=1):
            self._log(f"[{i}/{len(lines)}] {url}")
            res = self.download_url(url, with_avatars=with_avatars)
            if res.kind == "error":
                self._log(f"  ! {res.message}")
            results.append(res)
        return results
