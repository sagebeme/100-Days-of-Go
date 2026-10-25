// Day 68: talking to the Go API with fetch.

/** Route is what GET /api/routes returns, one per route. */
export interface Route {
  number?: string;
  operator: string;
  boardsAt: string;
  goesTo: string;
  fareKsh?: number;
}

/** API_URL is where the Go API runs. Day 75 explains VITE_ variables. */
export const API_URL: string = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

/** ApiError is a response that wasn't 2xx. message is the server's {"error": "..."} if it sent one. */
export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

/**
 * fetchRoutes calls GET {API_URL}/api/routes?q=QUERY (QUERY URL-encoded) and returns the routes.
 * It passes signal to fetch so the caller can cancel. On a non-2xx answer it throws an ApiError
 * with the status and the body's "error" field, or "HTTP STATUS" if the body has none.
 */
export async function fetchRoutes(query: string, signal?: AbortSignal): Promise<Route[]> {
  const res = await fetch(`${API_URL}/api/routes?q=${encodeURIComponent(query)}`, { signal });
  if (!res.ok) {
    // fetch only rejects when there's no response at all. A 500 is still a response.
    let message = `HTTP ${res.status}`;
    try {
      const body: unknown = await res.json();
      if (typeof body === "object" && body !== null && typeof (body as { error?: unknown }).error === "string") {
        message = (body as { error: string }).error;
      }
    } catch {
      // not JSON: keep "HTTP 500"
    }
    throw new ApiError(res.status, message);
  }
  const data: unknown = await res.json();
  if (!Array.isArray(data)) {
    throw new ApiError(res.status, "unexpected response from the server");
  }
  return data as Route[];
}
