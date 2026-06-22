import { useState } from "react";
import { Download, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DownloadURL } from "../../wailsjs/go/main/App";
import type { DownloadResult } from "@/types/api";

export function UrlPanel({
  output,
  hd,
  watermark,
}: {
  output: string;
  hd: boolean;
  watermark: boolean;
}) {
  const [url, setUrl] = useState("");
  const [loading, setLoading] = useState(false);
  const [last, setLast] = useState<DownloadResult | null>(null);

  const run = async () => {
    if (!url.trim()) {
      toast.error("Paste a TikTok URL first");
      return;
    }
    setLoading(true);
    setLast(null);
    try {
      const res = await DownloadURL({ url: url.trim(), output, hd, watermark });
      setLast(res);
      if (res.kind === "error") toast.error(res.message || "Download failed");
      else if (res.kind === "skipped") toast.info("Already downloaded");
      else toast.success(`Downloaded ${res.files.length} file(s)`);
    } catch (e) {
      toast.error(String(e));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-4">
      <div className="space-y-1.5">
        <Label>TikTok post URL</Label>
        <div className="flex gap-2">
          <Input
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && !loading && run()}
            placeholder="https://www.tiktok.com/@user/video/123…  (or vm.tiktok.com short link)"
          />
          <Button onClick={run} disabled={loading} className="no-drag">
            {loading ? <Loader2 className="size-4 animate-spin" /> : <Download className="size-4" />}
            Download
          </Button>
        </div>
      </div>

      {last && last.kind !== "error" && last.files.length > 0 && (
        <div className="rounded-lg border bg-card p-3 text-sm">
          <div className="mb-1 font-medium capitalize">{last.kind} · {last.media_id}</div>
          {last.files.map((f) => (
            <div key={f} className="truncate font-mono text-xs text-muted-foreground">{f}</div>
          ))}
        </div>
      )}
    </div>
  );
}
