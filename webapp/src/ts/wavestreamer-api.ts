const DEFAULT_HOST: string = window.location.host;

export default class WavestreamerApi {
  public readonly host: string;
  private eventListeners: Map<RadioEvent, UpdateEventListener[]> = new Map();
  private timeouts: Map<RadioEvent, number> = new Map();

  public constructor(host: string = DEFAULT_HOST) {
    this.host = host;

    // setup auto updates
    this.scheduleUpdate();
    document.addEventListener("visibilitychange", () => {
      if (document.visibilityState === "visible") {
        this.scheduleUpdate();
      }
    });
  }

  private get baseUrl(): string {
    return `http://${this.host}/api`;
  }

  public on(type: RadioEvent, listener: UpdateEventListener): void {
    if (!this.eventListeners.has(type)) {
      this.eventListeners.set(type, []);
    }

    const listeners = this.eventListeners.get(type) ?? [];
    listeners.push(listener);
  }

  private async request<T extends ApiBaseResponse>(
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

    const response = await fetch(this.baseUrl + path, init);
    const obj = (await response.json()) as ApiResponse;

    return obj.status === "ok"
      ? Promise.resolve(obj as T)
      : Promise.reject(obj as ApiErrorResponse);
  }

  public now(): Promise<ApiNowResponse> {
    return this.request("/now");
  }

  public async skip(): Promise<void> {
    await this.request("/skip", "PUT");
    this.scheduleUpdate();
  }

  public async pause(): Promise<void> {
    await this.request("/pause", "POST");
    this.scheduleUpdate();
  }

  public async repeat(): Promise<void> {
    await this.request("/repeat", "PUT");
  }

  public async search(query: string): Promise<ApiSearchResponse["results"]> {
    const nice_query = encodeURIComponent(query.trim());

    if (nice_query === "") {
      return Promise.resolve([]);
    }

    const response = await this.request<ApiSearchResponse>(
      `/library/search?query=${nice_query}`,
    );
    return response.results;
  }

  public async schedule(fileId: SearchResultEntry["id"]): Promise<void> {
    await this.request(
      "/schedule",
      "POST",
      new URLSearchParams({ file: fileId }),
    );
  }

  public async news(): Promise<void> {
    await this.request("/schedule/news", "POST");
  }

  public getConfig(): Promise<ApiConfigResponse> {
    return this.request<ApiConfigResponse>("/config", "POST");
  }

  public getDownloadUrl(clip: SearchResultEntry["id"]): string {
    return `${this.baseUrl}/library/download?file=${encodeURIComponent(clip)}`;
  }

  private async update(): Promise<void> {
    // clear timeout in case this was called manually
    const timeout = this.timeouts.get("update");
    if (timeout) {
      window.clearTimeout(timeout);
    }

    // update
    const data = await this.now();

    // notify listeners
    if (this.eventListeners.has("update")) {
      this.eventListeners.get("update")!.forEach((listener) => listener(data));
    }

    // schedule next update
    window.setTimeout(
      this.update.bind(this),
      document.visibilityState === "visible" ? 3141 : 6666,
    );
  }

  private scheduleUpdate(): void {
    // clear last timeout
    const timeout = this.timeouts.get("update");
    if (timeout) {
      window.clearTimeout(timeout);
    }

    window.setTimeout(this.update.bind(this), 10);
  }
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

type UpdateEventListener = (data: NowData) => unknown;

export type SearchResultEntry = {
  id: string;
  name: string;
};

type ApiConfigResponse = {
  status: "ok";
  news: boolean;
};
