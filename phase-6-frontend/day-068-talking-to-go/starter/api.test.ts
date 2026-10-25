// @vitest-environment node
import { afterEach, describe, expect, it, vi } from "vitest";
import { API_URL, ApiError, fetchRoutes } from "./api";

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("fetchRoutes", () => {
  it("calls the API with the query encoded and returns the routes", async () => {
    const fake = vi.fn(async () => jsonResponse(200, [{ operator: "MCA Sacco", boardsAt: "Railways", goesTo: "Ngong" }]));
    vi.stubGlobal("fetch", fake);
    const got = await fetchRoutes("ongata rongai");
    expect(got).toEqual([{ operator: "MCA Sacco", boardsAt: "Railways", goesTo: "Ngong" }]);
    expect(fake).toHaveBeenCalledTimes(1);
    const [url] = fake.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe(`${API_URL}/api/routes?q=ongata%20rongai`);
  });
  it("passes the abort signal to fetch", async () => {
    const fake = vi.fn(async () => jsonResponse(200, []));
    vi.stubGlobal("fetch", fake);
    const controller = new AbortController();
    await fetchRoutes("", controller.signal);
    const [, init] = fake.mock.calls[0] as unknown as [string, RequestInit];
    expect(init?.signal).toBe(controller.signal);
  });
  const failures: { name: string; res: () => Response; status: number; message: string }[] = [
    { name: "a 500 with an error body", res: () => jsonResponse(500, { error: "the route database is down" }), status: 500, message: "the route database is down" },
    { name: "a 404 without JSON", res: () => new Response("not found", { status: 404 }), status: 404, message: "HTTP 404" },
  ];
  for (const c of failures) {
    it(`throws an ApiError for ${c.name}`, async () => {
      vi.stubGlobal("fetch", vi.fn(async () => c.res()));
      const err = await fetchRoutes("x").catch((e: unknown) => e);
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).status).toBe(c.status);
      expect((err as ApiError).message).toBe(c.message);
    });
  }
});
