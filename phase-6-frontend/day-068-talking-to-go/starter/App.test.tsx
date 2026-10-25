import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { RouteSearch } from "./App";
import type { Route } from "./api";

const NGONG: Route = { number: "111", operator: "MCA Sacco", boardsAt: "Railways", goesTo: "Ngong", fareKsh: 100 };
const KIBERA: Route = { operator: "Highrise Kibera Sacco", boardsAt: "Railways", goesTo: "Kibera" };

function ok(body: unknown): Response {
  return new Response(JSON.stringify(body), { status: 200 });
}

/** pending returns a fetch stand-in whose answers you release by hand, so a test controls the timing. */
function pending() {
  const calls: { url: string; signal?: AbortSignal; resolve: (r: Response) => void }[] = [];
  const fake = vi.fn((url: string, init?: RequestInit) => {
    return new Promise<Response>((resolve, reject) => {
      const signal = init?.signal ?? undefined;
      signal?.addEventListener("abort", () => reject(new DOMException("aborted", "AbortError")));
      calls.push({ url, signal, resolve });
    });
  });
  return { fake, calls };
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("RouteSearch", () => {
  it("shows Loading, then the routes", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ok([NGONG, KIBERA])));
    render(<RouteSearch />);
    expect(screen.getByRole("status").textContent).toBe("Loading routes...");
    const list = await screen.findByRole("list", { name: "Routes" });
    const items = within(list).getAllByRole("listitem");
    expect(items.map((li) => li.textContent)).toEqual(["111 to Ngong, Ksh100", "To Kibera, ask the conductor"]);
    expect(screen.queryByRole("status")).toBeNull();
  });
  it("says so when nothing matches", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ok([])));
    render(<RouteSearch />);
    await screen.findByText("No routes match.");
  });
  it("shows the server's error and retries", async () => {
    const fake = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: "the route database is down" }), { status: 500 }))
      .mockResolvedValueOnce(ok([NGONG]));
    vi.stubGlobal("fetch", fake);
    render(<RouteSearch />);
    const alert = await screen.findByRole("alert");
    expect(alert.textContent).toContain("Couldn't load routes: the route database is down");
    await userEvent.setup().click(within(alert).getByRole("button", { name: "Try again" }));
    await screen.findByText("111 to Ngong, Ksh100");
    expect(screen.queryByRole("alert")).toBeNull();
  });
  it("shows a network failure as an error, not a crash", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Promise.reject(new TypeError("Failed to fetch"))));
    render(<RouteSearch />);
    const alert = await screen.findByRole("alert");
    expect(alert.textContent).toContain("Couldn't load routes: Failed to fetch");
  });
  it("searches as you type, aborting the request it no longer needs", async () => {
    const { fake, calls } = pending();
    vi.stubGlobal("fetch", fake);
    const user = userEvent.setup();
    render(<RouteSearch />);
    await user.type(screen.getByLabelText("Destination"), "ng");
    // One request on load, then one per keystroke.
    expect(calls.map((c) => new URL(c.url).searchParams.get("q"))).toEqual(["", "n", "ng"]);
    expect(calls[0].signal?.aborted, "the first request was aborted").toBe(true);
    expect(calls[1].signal?.aborted, "the second request was aborted").toBe(true);
    expect(calls[2].signal?.aborted, "the latest request is still live").toBe(false);
    calls[2].resolve(ok([NGONG]));
    await screen.findByText("111 to Ngong, Ksh100");
    expect(screen.queryByRole("alert"), "an abort is not an error").toBeNull();
  });
});
