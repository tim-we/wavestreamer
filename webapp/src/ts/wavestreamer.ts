const DEFAULT_HOST: string = window.location.host;

export default class Wavestreamer {
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

    public get base_url(): string {
        return `http://${this.host}/api/v1.0`;
    }

    public on(type: RadioEvent, listener: UpdateEventListener): void {
        if (!this.eventListeners.has(type)) {
            this.eventListeners.set(type, []);
        }

        const listeners = this.eventListeners.get(type) ?? [];
        listeners.push(listener);
    }

    public async api_request<T extends APIBaseResponse>(
        path: string,
        method: HTTPMethod = "GET",
        data: any = null
    ): Promise<T> {
        let init: RequestInitData = {
            method: method,
            cache: "no-store",
            follow: "error",
            body: data,
        };

        if (data === null) {
            delete init.body;
        }

        let response = await fetch(this.base_url + path, init);
        let obj = (await response.json()) as APIResponse;

        return obj.status === "ok"
            ? Promise.resolve(obj as T)
            : Promise.reject(obj as APIErrorResponse);
    }

    public now(): Promise<APINowResponse> {
        return this.api_request("/now");
    }

    public async skip(): Promise<void> {
        await this.api_request("/skip", "PUT");
        this.scheduleUpdate();
    }

    public async pause(): Promise<void> {
        await this.api_request("/pause", "POST");
        this.scheduleUpdate();
    }

    public async repeat(): Promise<void> {
        await this.api_request("/repeat", "PUT");
    }

    public async extensions() {
        const obj = await this.api_request<APIExtensionResponse>("/extensions");
        return obj.extensions;
    }

    public async search(query: string): Promise<string[]> {
        const nice_query = encodeURIComponent(query.trim());

        if (nice_query === "") {
            return Promise.resolve([]);
        }

        const response = await this.api_request<APISearchResponse>(
            "/library/search?query=" + nice_query
        );
        return response.results;
    }

    public async schedule(clip: string): Promise<void> {
        const form = new FormData();
        form.append("file", clip);
        // @ts-ignore
        await this.api_request("/schedule", "POST", new URLSearchParams(form));
    }

    public download_url(clip: string): string {
        return `${this.base_url}/library/download?file=${encodeURIComponent(
            clip
        )}`;
    }

    private async update(): Promise<void> {
        // clear timeout in case this was called manually
        const timeout = this.timeouts.get("update");
        if (timeout) {
            window.clearTimeout(timeout);
        }

        // update
        let data = await this.now();

        // notify listeners
        if (this.eventListeners.has("update")) {
            this.eventListeners
                .get("update")!
                .forEach((listener) => listener(data));
        }

        // schedule next update
        window.setTimeout(
            this.update.bind(this),
            document.visibilityState === "visible" ? 3141 : 6666
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

type APIErrorResponse = { status: "error"; message: string };

type APIBaseResponse = {
    status: "ok";
};

type APIResponse = APIBaseResponse | APIErrorResponse;

type APINowResponse = APIBaseResponse & NowData;

type APIExtensionResponse = {
    status: "ok";
    extensions: { name: string; command: string }[];
};

type APISearchResponse = {
    status: "ok";
    results: string[];
};

type NowData = {
    current: string;
    history: HistoryEntry[];
    library: { hosts: number; music: number; other: number };
};

export type HistoryEntry = {
    start: string;
    title: string;
    skipped: boolean;
    userScheduled: boolean;
}

type UpdateEventListener = (data: NowData) => any;
