"""File downloading, on-disk layout, and dedup index.

Reproduces the C# app's storage scheme:

    <root>/<username>/Videos/<id>.mp4
    <root>/<username>/Images/<id>_<n>.jpeg
    <root>/<username>/Avatars/<username>_<n>.jpeg
    <root>/<username>/<username>_index.txt   (one downloaded key per line)
"""

from __future__ import annotations

import time
from pathlib import Path

import requests

_DL_HEADERS = {
    "User-Agent": (
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
        "(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
    )
}

_CHUNK = 1 << 16  # 64 KiB


def _safe_username(username: str | None) -> str:
    name = (username or "unknown").strip().lstrip("@")
    return name or "unknown"


class Storage:
    """Per-user folders plus the dedup index file."""

    def __init__(self, root: str | Path, username: str | None):
        self.username = _safe_username(username)
        self.user_dir = Path(root) / self.username
        self.videos_dir = self.user_dir / "Videos"
        self.images_dir = self.user_dir / "Images"
        self.avatars_dir = self.user_dir / "Avatars"
        self.index_path = self.user_dir / f"{self.username}_index.txt"
        self._index: set[str] | None = None

    # --- dedup index ----------------------------------------------------
    def _load_index(self) -> set[str]:
        if self._index is None:
            if self.index_path.exists():
                self._index = {
                    line.strip()
                    for line in self.index_path.read_text(
                        encoding="utf-8", errors="ignore"
                    ).splitlines()
                    if line.strip()
                }
            else:
                self._index = set()
        return self._index

    def has(self, key: str) -> bool:
        return key in self._load_index()

    def mark(self, key: str) -> None:
        self._load_index().add(key)
        self.user_dir.mkdir(parents=True, exist_ok=True)
        with self.index_path.open("a", encoding="utf-8") as fh:
            fh.write(f"{key}\n")


def download_file(
    url: str,
    dest: str | Path,
    *,
    session: requests.Session | None = None,
    retries: int = 5,
    timeout: int = 60,
) -> Path:
    """Stream a URL to disk with retries, like DownloadVideoWithBufferedWrite()."""
    sess = session or requests.Session()
    dest = Path(dest)
    dest.parent.mkdir(parents=True, exist_ok=True)

    last_exc: Exception | None = None
    for attempt in range(retries):
        try:
            with sess.get(
                url, headers=_DL_HEADERS, stream=True, timeout=timeout
            ) as resp:
                resp.raise_for_status()
                tmp = dest.with_suffix(dest.suffix + ".part")
                with tmp.open("wb") as fh:
                    for chunk in resp.iter_content(chunk_size=_CHUNK):
                        if chunk:
                            fh.write(chunk)
                tmp.replace(dest)
            return dest
        except (requests.RequestException, OSError) as exc:  # noqa: PERF203
            last_exc = exc
            if attempt < retries - 1:
                time.sleep(1.5)
    raise RuntimeError(f"Failed to download {url} after {retries} attempts: {last_exc}")
