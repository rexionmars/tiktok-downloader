import { useEffect, useState } from "react";
import { TitleBar } from "@/components/TitleBar";
import { Sidebar, type PageType } from "@/components/Sidebar";
import { OutputBar } from "@/components/OutputBar";
import { UrlPanel } from "@/components/UrlPanel";
import { BatchPanel } from "@/components/BatchPanel";
import { ProfilePanel } from "@/components/ProfilePanel";
import { HistoryPanel } from "@/components/HistoryPanel";
import { GetDefaultOutput } from "../wailsjs/go/main/App";

const TITLES: Record<PageType, string> = {
  url: "Download a single post",
  batch: "Batch download",
  profile: "Download a whole profile",
  history: "History",
};

function App() {
  const [page, setPage] = useState<PageType>("url");
  const [output, setOutput] = useState("");
  const [hd, setHd] = useState(false);
  const [watermark, setWatermark] = useState(false);

  useEffect(() => {
    GetDefaultOutput().then(setOutput).catch(() => {});
  }, []);

  return (
    <div className="h-screen overflow-hidden bg-background text-foreground">
      <TitleBar />
      <Sidebar current={page} onChange={setPage} />

      <div className="fixed top-10 right-0 bottom-0 left-14 overflow-y-auto">
        <div className="mx-auto max-w-3xl space-y-6 p-6">
          <h1 className="text-lg font-semibold">{TITLES[page]}</h1>

          {page !== "history" && (
            <OutputBar
              output={output}
              onOutputChange={setOutput}
              hd={hd}
              onHdChange={setHd}
              watermark={watermark}
              onWatermarkChange={setWatermark}
            />
          )}

          {page === "url" && <UrlPanel output={output} hd={hd} watermark={watermark} />}
          {page === "batch" && <BatchPanel output={output} hd={hd} watermark={watermark} />}
          {page === "profile" && <ProfilePanel output={output} hd={hd} watermark={watermark} />}
          {page === "history" && <HistoryPanel />}
        </div>
      </div>
    </div>
  );
}

export default App;
