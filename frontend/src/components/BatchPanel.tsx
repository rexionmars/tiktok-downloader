import { useEffect, useRef, useState } from "react";
import { Download, Loader2, Square, CheckCircle2, XCircle, SkipForward } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { DownloadBatch, StopBatch } from "../../wailsjs/go/main/App";
import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime";
import type { BatchProgress } from "@/types/api";

interface Row {
  url: string;
  status: "pending" | "done" | "skipped" | "error";
  message?: string;
}

export function BatchPanel({
  output,
  hd,
  watermark,
}: {
  output: string;
  hd: boolean;
  watermark: boolean;
}) {
  const [text, setText] = useState("");
  const [rows, setRows] = useState<Row[]>([]);
  const [running, setRunning] = useState(false);
  const [progress, setProgress] = useState(0);
  const rowsRef = useRef<Row[]>([]);

  useEffect(() => {
    EventsOn("batch:progress", (p: BatchProgress) => {
      setProgress(Math.round((p.index / p.total) * 100));
      const status: Row["status"] =
        p.result.kind === "error"
          ? "error"
          : p.result.kind === "skipped"
            ? "skipped"
            : "done";
      rowsRef.current = rowsRef.current.map((r) =>
        r.url === p.url ? { ...r, status, message: p.result.message } : r
      );
      setRows([...rowsRef.current]);
    });
    return () => EventsOff("batch:progress");
  }, []);

  const run = async () => {
    const urls = text
      .split("\n")
      .map((l) => l.trim())
      .filter(Boolean);
    if (urls.length === 0) {
      toast.error("Paste at least one URL");
      return;
    }
    const initial: Row[] = urls.map((url) => ({ url, status: "pending" }));
    rowsRef.current = initial;
    setRows(initial);
    setProgress(0);
    setRunning(true);
    try {
      await DownloadBatch({ urls, output, hd, watermark });
      const ok = rowsRef.current.filter((r) => r.status === "done").length;
      toast.success(`Batch finished — ${ok}/${urls.length} downloaded`);
    } catch (e) {
      toast.error(String(e));
    } finally {
      setRunning(false);
    }
  };

  const icon = (s: Row["status"]) =>
    s === "done" ? (
      <CheckCircle2 className="size-4 text-green-500" />
    ) : s === "skipped" ? (
      <SkipForward className="size-4 text-muted-foreground" />
    ) : s === "error" ? (
      <XCircle className="size-4 text-destructive" />
    ) : (
      <Loader2 className="size-4 animate-spin text-muted-foreground" />
    );

  return (
    <div className="space-y-4">
      <div className="space-y-1.5">
        <Label>URLs (one per line)</Label>
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          rows={6}
          placeholder={"https://www.tiktok.com/@user/video/111\nhttps://www.tiktok.com/@user/photo/222"}
          className="no-drag w-full resize-y rounded-md border bg-transparent p-3 font-mono text-xs outline-none focus-visible:ring-ring/50 focus-visible:ring-[3px] dark:bg-input/30"
        />
      </div>

      <div className="flex gap-2">
        <Button onClick={run} disabled={running} className="no-drag">
          {running ? <Loader2 className="size-4 animate-spin" /> : <Download className="size-4" />}
          Download all
        </Button>
        {running && (
          <Button variant="destructive" onClick={() => StopBatch()} className="no-drag">
            <Square className="size-4" /> Stop
          </Button>
        )}
      </div>

      {rows.length > 0 && (
        <>
          <Progress value={progress} />
          <div className="max-h-64 space-y-1 overflow-y-auto rounded-lg border bg-card p-2">
            {rows.map((r) => (
              <div key={r.url} className="flex items-center gap-2 px-1 py-1 text-xs">
                {icon(r.status)}
                <span className="truncate font-mono text-muted-foreground">{r.url}</span>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
