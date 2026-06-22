import { Folder, FolderOpen } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { SelectOutputFolder, OpenFolder } from "../../wailsjs/go/main/App";

export function OutputBar({
  output,
  onOutputChange,
  hd,
  onHdChange,
  watermark,
  onWatermarkChange,
}: {
  output: string;
  onOutputChange: (v: string) => void;
  hd: boolean;
  onHdChange: (v: boolean) => void;
  watermark: boolean;
  onWatermarkChange: (v: boolean) => void;
}) {
  const pick = async () => {
    const dir = await SelectOutputFolder();
    if (dir) onOutputChange(dir);
  };

  return (
    <div className="space-y-3 rounded-lg border bg-card p-3">
      <div className="space-y-1.5">
        <Label className="text-xs text-muted-foreground">Download folder</Label>
        <div className="flex gap-2">
          <Input
            value={output}
            onChange={(e) => onOutputChange(e.target.value)}
            placeholder="Choose a folder…"
            className="font-mono text-xs"
          />
          <Button variant="outline" size="icon" onClick={pick} title="Browse">
            <Folder className="size-4" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            onClick={() => output && OpenFolder(output)}
            title="Open folder"
          >
            <FolderOpen className="size-4" />
          </Button>
        </div>
      </div>
      <div className="flex items-center gap-6">
        <Label className="gap-2">
          <Switch checked={hd} onCheckedChange={onHdChange} />
          HD (tikwm.com)
        </Label>
        <Label className="gap-2">
          <Switch
            checked={watermark}
            onCheckedChange={onWatermarkChange}
            disabled={hd}
          />
          Watermark
        </Label>
      </div>
    </div>
  );
}
