import { useState } from "react";
import { Download, Loader2, Search, Film, Image as ImageIcon } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrapeProfile, DownloadProfile } from "../../wailsjs/go/main/App";
import { BrowserOpenURL } from "../../wailsjs/runtime/runtime";
import type { PostPreview } from "@/types/api";

export function ProfilePanel({
  output,
  hd,
  watermark,
}: {
  output: string;
  hd: boolean;
  watermark: boolean;
}) {
  const [username, setUsername] = useState("");
  const [scraping, setScraping] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [posts, setPosts] = useState<PostPreview[]>([]);

  const scan = async () => {
    if (!username.trim()) return toast.error("Enter a username");
    setScraping(true);
    setPosts([]);
    try {
      toast.info("Opening browser to scan the profile…");
      const found = (await ScrapeProfile(username.trim())) ?? [];
      setPosts(found);
      toast.success(`Found ${found.length} posts`);
    } catch (e) {
      toast.error(String(e));
    } finally {
      setScraping(false);
    }
  };

  const run = async () => {
    if (!username.trim()) return toast.error("Enter a username");
    setDownloading(true);
    try {
      const results =
        (await DownloadProfile({ username: username.trim(), output, hd, watermark })) ?? [];
      const ok = results.filter((r) => r.kind === "video" || r.kind === "image").length;
      toast.success(`Profile done — ${ok}/${results.length} downloaded`);
    } catch (e) {
      toast.error(String(e));
    } finally {
      setDownloading(false);
    }
  };

  const busy = scraping || downloading;

  return (
    <div className="space-y-4">
      <div className="space-y-1.5">
        <Label>Username</Label>
        <div className="flex gap-2">
          <span className="flex h-9 items-center rounded-md border bg-muted px-3 text-sm text-muted-foreground">@</span>
          <Input
            value={username}
            onChange={(e) => setUsername(e.target.value.replace(/^@/, ""))}
            onKeyDown={(e) => e.key === "Enter" && !busy && scan()}
            placeholder="someuser"
          />
          <Button variant="outline" onClick={scan} disabled={busy} className="no-drag">
            {scraping ? <Loader2 className="size-4 animate-spin" /> : <Search className="size-4" />}
            Scan
          </Button>
          <Button onClick={run} disabled={busy} className="no-drag">
            {downloading ? <Loader2 className="size-4 animate-spin" /> : <Download className="size-4" />}
            Download all
          </Button>
        </div>
      </div>

      {posts.length > 0 && (
        <>
          <div className="text-sm">
            <span className="font-medium">{posts.length}</span> posts found for{" "}
            <span className="font-mono">@{username}</span>
          </div>
          <div className="grid grid-cols-4 gap-2 sm:grid-cols-5 md:grid-cols-6">
            {posts.map((p) => (
              <button
                key={p.url}
                onClick={() => BrowserOpenURL(p.url)}
                title={p.url}
                className="no-drag group relative aspect-[3/4] overflow-hidden rounded-md border bg-muted transition-transform hover:scale-[1.03]"
              >
                {p.thumbnail ? (
                  <img
                    src={p.thumbnail}
                    alt=""
                    loading="lazy"
                    referrerPolicy="no-referrer"
                    className="size-full object-cover"
                  />
                ) : (
                  <div className="flex size-full items-center justify-center text-muted-foreground">
                    {p.kind === "photo" ? (
                      <ImageIcon className="size-5" />
                    ) : (
                      <Film className="size-5" />
                    )}
                  </div>
                )}
                <span className="absolute right-1 top-1 rounded bg-black/60 p-0.5 text-white">
                  {p.kind === "photo" ? (
                    <ImageIcon className="size-3" />
                  ) : (
                    <Film className="size-3" />
                  )}
                </span>
              </button>
            ))}
          </div>
        </>
      )}

      <p className="text-xs text-muted-foreground">
        Scanning opens a visible Chrome window to scroll the profile and collect
        every video/photo link with a thumbnail, then "Download all" fetches them
        with the options above.
      </p>
    </div>
  );
}
