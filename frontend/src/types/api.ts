// Mirrors the Go backend.Result and app.go request/response types.

// `kind` is one of "video" | "image" | "skipped" | "error", but is typed as
// string to stay compatible with the Wails-generated backend.Result type
// (Go has no literal-union types).
export type ResultKind = "video" | "image" | "skipped" | "error";

export interface DownloadResult {
  media_id: string;
  kind: string;
  files: string[];
  message: string;
}

export interface BatchProgress {
  index: number;
  total: number;
  url: string;
  result: DownloadResult;
}

export interface PostPreview {
  url: string;
  thumbnail: string;
  kind: string; // "video" | "photo"
}

export interface HistoryItem {
  url: string;
  media_id: string;
  kind: string;
  timestamp: number;
}

export interface DownloadRequest {
  url: string;
  output: string;
  hd: boolean;
  watermark: boolean;
}

export interface BatchRequest {
  urls: string[];
  output: string;
  hd: boolean;
  watermark: boolean;
}

export interface ProfileRequest {
  username: string;
  output: string;
  hd: boolean;
  watermark: boolean;
}
