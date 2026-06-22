import { Minus, X, Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { WindowMinimise, Quit } from "../../wailsjs/runtime/runtime";
import { Button } from "@/components/ui/button";

export function TitleBar() {
  const { theme, setTheme } = useTheme();
  return (
    <div className="draggable fixed top-0 right-0 left-0 z-50 flex h-10 items-center justify-between border-b bg-background px-3">
      <div className="flex items-center gap-2 text-sm font-medium">
        <span className="text-base">⬇️</span>
        TikTok Downloader
      </div>
      <div className="no-drag flex items-center gap-1">
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
          aria-label="Toggle theme"
        >
          {theme === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />}
        </Button>
        <Button variant="ghost" size="icon-sm" onClick={() => WindowMinimise()} aria-label="Minimize">
          <Minus className="size-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={() => Quit()}
          aria-label="Close"
          className="hover:bg-destructive hover:text-white"
        >
          <X className="size-4" />
        </Button>
      </div>
    </div>
  );
}
