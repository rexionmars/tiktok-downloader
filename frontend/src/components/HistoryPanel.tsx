import { useEffect, useState } from "react";
import { Trash2, Film, Image as ImageIcon } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { GetHistory, ClearHistory } from "../../wailsjs/go/main/App";
import type { HistoryItem } from "@/types/api";

export function HistoryPanel() {
  const [items, setItems] = useState<HistoryItem[]>([]);

  const load = async () => setItems((await GetHistory()) ?? []);
  useEffect(() => {
    load();
  }, []);

  const clear = async () => {
    await ClearHistory();
    setItems([]);
    toast.info("History cleared");
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-medium">Recent downloads</h2>
        {items.length > 0 && (
          <Button variant="outline" size="sm" onClick={clear} className="no-drag">
            <Trash2 className="size-4" /> Clear
          </Button>
        )}
      </div>

      {items.length === 0 ? (
        <p className="text-sm text-muted-foreground">No downloads yet.</p>
      ) : (
        <div className="space-y-1 rounded-lg border bg-card p-2">
          {items.map((it, i) => (
            <div key={`${it.media_id}-${i}`} className="flex items-center gap-3 px-1 py-1.5 text-sm">
              {it.kind === "image" ? (
                <ImageIcon className="size-4 text-muted-foreground" />
              ) : (
                <Film className="size-4 text-muted-foreground" />
              )}
              <span className="font-mono text-xs">{it.media_id}</span>
              <span className="ml-auto text-xs text-muted-foreground">
                {new Date(it.timestamp * 1000).toLocaleString()}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
