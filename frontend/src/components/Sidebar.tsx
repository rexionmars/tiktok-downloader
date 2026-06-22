import { Link2, ListOrdered, User, History } from "lucide-react";
import { cn } from "@/lib/utils";

export type PageType = "url" | "batch" | "profile" | "history";

const ITEMS: { id: PageType; label: string; icon: typeof Link2 }[] = [
  { id: "url", label: "Single URL", icon: Link2 },
  { id: "batch", label: "Batch", icon: ListOrdered },
  { id: "profile", label: "Profile", icon: User },
  { id: "history", label: "History", icon: History },
];

export function Sidebar({
  current,
  onChange,
}: {
  current: PageType;
  onChange: (p: PageType) => void;
}) {
  return (
    <div className="fixed top-10 bottom-0 left-0 flex w-14 flex-col items-center gap-1 border-r bg-background py-3">
      {ITEMS.map(({ id, label, icon: Icon }) => (
        <button
          key={id}
          onClick={() => onChange(id)}
          title={label}
          className={cn(
            "no-drag flex size-10 items-center justify-center rounded-md transition-colors",
            current === id
              ? "bg-primary text-primary-foreground"
              : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          )}
        >
          <Icon className="size-5" />
        </button>
      ))}
    </div>
  );
}
