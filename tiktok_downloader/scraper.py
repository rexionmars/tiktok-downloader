"""Profile scraping via Playwright (mass download by username).

Ports the Playwright automation from MainForm.cs:
- open https://www.tiktok.com/@<username>
- scroll to the bottom repeatedly until the page height stops growing
- collect every <a> whose href contains /video/ or /photo/
- write them to <root>/<username>_combined_links.txt

Requires the optional `playwright` extra:
    pip install playwright && python -m playwright install chromium
"""

from __future__ import annotations

from pathlib import Path

_COLLECT_JS = """
() => {
  const urls = new Set();
  document.querySelectorAll('a').forEach(a => {
    if (a.href.includes('/video/') || a.href.includes('/photo/')) {
      urls.add(a.href);
    }
  });
  return Array.from(urls);
}
"""


def scrape_user_links(
    username: str,
    root: str,
    *,
    headless: bool = False,
    scroll_wait_ms: int = 10000,
    max_scrolls: int = 200,
    executable_path: str | None = None,
) -> Path:
    """Scrape all post links for a user and write them to a .txt file.

    Returns the path to <username>_combined_links.txt. Raises ImportError if
    Playwright isn't installed.
    """
    try:
        from playwright.sync_api import sync_playwright
    except ImportError as exc:  # pragma: no cover
        raise ImportError(
            "Playwright is required for username scraping. Install with:\n"
            "    pip install playwright\n"
            "    python -m playwright install chromium"
        ) from exc

    username = username.strip().lstrip("@")
    base_url = f"https://www.tiktok.com/@{username}"

    with sync_playwright() as pw:
        launch_kwargs: dict = {"headless": headless}
        if executable_path:
            launch_kwargs["executable_path"] = executable_path
        browser = pw.chromium.launch(**launch_kwargs)
        try:
            page = browser.new_context().new_page()
            page.goto(base_url, timeout=120000)

            # Scroll until the page stops growing (same loop as the C# app).
            for _ in range(max_scrolls):
                initial = page.evaluate("() => document.body.scrollHeight")
                page.evaluate("window.scrollTo(0, document.body.scrollHeight)")
                page.wait_for_timeout(scroll_wait_ms)
                new_height = page.evaluate("() => document.body.scrollHeight")
                if new_height == initial:
                    break

            links = page.evaluate(_COLLECT_JS)
        finally:
            browser.close()

    # Keep only this user's posts and dedupe while preserving order.
    seen: set[str] = set()
    filtered: list[str] = []
    for url in links:
        if f"/@{username}/" in url or f"/{username}/" in url:
            if url not in seen:
                seen.add(url)
                filtered.append(url)

    out_dir = Path(root)
    out_dir.mkdir(parents=True, exist_ok=True)
    out_path = out_dir / f"{username}_combined_links.txt"
    out_path.write_text("\n".join(filtered) + ("\n" if filtered else ""), encoding="utf-8")
    return out_path
