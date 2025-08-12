import { signal } from "@preact/signals";

const baseUrl = `http://${window.location.host}/api`;

export const nowDataSignal = signal<NowData | null>(null);

const timeouts: Map<RadioEvent, number> = new Map();

export function init() {
  // setup auto updates
  scheduleUpdate();
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible") {
      scheduleUpdate();
    }
  });
}

export function now(): Promise<ApiNowResponse> {
  return request("/now");
}

export async function skip(): Promise<void> {
  await request("/skip", "PUT");
  scheduleUpdate();
}

export async function pause(): Promise<void> {
  await request("/pause", "POST");
  scheduleUpdate();
}

export async function repeat(): Promise<void> {
  await request("/repeat", "PUT");
}

export async function search(
  query: string,
): Promise<ApiSearchResponse["results"]> {
  const nice_query = encodeURIComponent(query.trim());

  if (nice_query === "") {
    return Promise.resolve([]);
  }

  const response = await request<ApiSearchResponse>(
    `/library/search?query=${nice_query}`,
  );
  return response.results;
}

export async function schedule(fileId: SearchResultEntry["id"]): Promise<void> {
  await request("/schedule", "POST", new URLSearchParams({ file: fileId }));
}

export async function news(): Promise<void> {
  await request("/schedule/news", "POST");
}

async function request<T extends ApiBaseResponse>(
  path: string,
  method: HTTPMethod = "GET",
  data?: URLSearchParams | FormData,
): Promise<T> {
  const init: RequestInitData = {
    method: method,
    cache: "no-store",
    follow: "error",
    body: data,
  };

  if (data === null) {
    // biome-ignore lint/performance/noDelete: Performance is not critical here
    delete init.body;
  }

  const response = await fetch(baseUrl + path, init);
  const obj = (await response.json()) as ApiResponse;

  return obj.status === "ok"
    ? Promise.resolve(obj as T)
    : Promise.reject(obj as ApiErrorResponse);
}

function scheduleUpdate(): void {
  // clear last timeout
  const timeout = timeouts.get("update");
  if (timeout) {
    window.clearTimeout(timeout);
  }

  window.setTimeout(update, 10);
}

export function getConfig(): Promise<ApiConfigResponse> {
  return request<ApiConfigResponse>("/config", "POST");
}

export function getDownloadUrl(clip: SearchResultEntry["id"]): string {
  return `${baseUrl}/library/download?file=${encodeURIComponent(clip)}`;
}

async function update(): Promise<void> {
  // clear timeout in case this was called manually
  const timeout = timeouts.get("update");
  if (timeout) {
    window.clearTimeout(timeout);
  }

  // update
  const data = await now();

  // notify listeners
  nowDataSignal.value = data;

  // schedule next update
  window.setTimeout(
    update.bind(this),
    document.visibilityState === "visible" ? 3141 : 6666,
  );
}

type HTTPMethod = "GET" | "POST" | "PUT" | "DELETE";

type RadioEvent = "update";

type RequestInitData = RequestInit & { follow: "error" };

type ApiErrorResponse = { status: "error"; message: string };

type ApiBaseResponse = {
  status: "ok";
};

type ApiResponse = ApiBaseResponse | ApiErrorResponse;

type ApiNowResponse = ApiBaseResponse & NowData;

type ApiSearchResponse = {
  status: "ok";
  results: SearchResultEntry[];
};

type NowData = {
  current: string;
  isPause: boolean;
  history: HistoryEntry[];
  library: { hosts: number; music: number; other: number };
  uptime: string;
};

export type HistoryEntry = {
  start: string;
  title: string;
  skipped: boolean;
  userScheduled: boolean;
};

export type SearchResultEntry = {
  id: string;
  name: string;
};

type ApiConfigResponse = {
  status: "ok";
  news: boolean;
};
