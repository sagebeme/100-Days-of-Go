// A fake fetch for tests: you list the requests it knows, it answers them and remembers every call.

/** Handler answers one fake request. */
export type Handler = (req: Request) => Response | Promise<Response>;

export interface FakeFetch {
  /** fetch has the same shape as the real one, so vi.stubGlobal("fetch", fake.fetch) works. */
  fetch: (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;
  /** calls is every request made, in order, as "METHOD /path" (no origin, no query string). */
  calls: string[];
}

/** json makes a Response with a JSON body and Content-Type: application/json. */
export function json(status: number, body: unknown): Response {
  // Your code here
  return new Response(null, { status: 500 });
}

/**
 * createFakeFetch answers each request with the handler whose key is "METHOD /path", e.g.
 * "GET /api/trips". The method is upper case, the path is the URL's pathname. Each call is
 * recorded in calls before the handler runs. A request with no handler gets
 * json(404, { error: "no fake for METHOD /path" }), so a missing fake is easy to spot.
 * Handlers can be async. A handler that throws makes fetch reject, like a network failure.
 */
export function createFakeFetch(routes: Record<string, Handler>): FakeFetch {
  // Your code here
  return { fetch: async () => new Response(null, { status: 500 }), calls: [] };
}
