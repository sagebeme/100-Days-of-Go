import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { API_URL, Auth, LoginError, SessionExpiredError } from "./auth";

/**
 * fakeServer stands in for the Go API. It hands out tokens tok1, tok2, ... and accepts only the
 * newest one. expire() makes the current access token stop working; killRefresh() logs the cookie out.
 */
function fakeServer() {
  let n = 0;
  let current = "";
  let refreshOk = false;
  const calls: { path: string; auth: string | null; credentials?: RequestCredentials }[] = [];
  const json = (status: number, body: unknown) => new Response(JSON.stringify(body), { status });
  const user = { phone: "0712345678", name: "Amina" };
  const issue = () => {
    current = `tok${++n}`;
    refreshOk = true;
    return json(200, { accessToken: current, user });
  };
  const fake = vi.fn(async (url: string, init: RequestInit = {}) => {
    const path = url.replace(API_URL, "");
    const auth = new Headers(init.headers).get("Authorization");
    calls.push({ path, auth, credentials: init.credentials });
    await Promise.resolve(); // answer later, like a real network
    switch (path) {
      case "/api/login": {
        const body = JSON.parse(init.body as string);
        return body.password === "matatu-2026" ? issue() : json(401, { error: "wrong phone number or password" });
      }
      case "/api/refresh":
        return refreshOk ? issue() : json(401, { error: "please log in" });
      case "/api/logout":
        refreshOk = false;
        return new Response(null, { status: 204 });
      default:
        return auth === `Bearer ${current}` && current !== "" ? json(200, [{ reference: "K7Q2M9XA" }]) : json(401, {});
    }
  });
  return {
    fake,
    calls,
    expire: () => {
      current = "expired";
    },
    killRefresh: () => {
      refreshOk = false;
    },
  };
}

let server: ReturnType<typeof fakeServer>;
beforeEach(() => {
  server = fakeServer();
  vi.stubGlobal("fetch", server.fake);
  localStorage.clear();
  sessionStorage.clear();
});
afterEach(() => {
  vi.unstubAllGlobals();
});

describe("Auth", () => {
  it("logs in, keeps the token in memory only, and sends cookies", async () => {
    const auth = new Auth();
    const user = await auth.login("0712345678", "matatu-2026");
    expect(user).toEqual({ phone: "0712345678", name: "Amina" });
    expect(auth.loggedIn).toBe(true);
    expect(localStorage.length, "nothing in localStorage").toBe(0);
    expect(sessionStorage.length, "nothing in sessionStorage").toBe(0);
    expect(server.calls[0].credentials, "login sends credentials").toBe("include");
  });

  it("throws LoginError for a wrong password", async () => {
    const auth = new Auth();
    await expect(auth.login("0712345678", "1234")).rejects.toBeInstanceOf(LoginError);
    expect(auth.loggedIn).toBe(false);
  });

  it("adds the access token to API calls", async () => {
    const auth = new Auth();
    await auth.login("0712345678", "matatu-2026");
    const res = await auth.apiFetch("/api/me/bookings");
    expect(res.status).toBe(200);
    expect(server.calls.at(-1)?.auth).toBe("Bearer tok1");
  });

  it("refreshes once on a 401 and retries with the new token", async () => {
    const auth = new Auth();
    await auth.login("0712345678", "matatu-2026");
    server.expire();
    const res = await auth.apiFetch("/api/me/bookings");
    expect(res.status).toBe(200);
    expect(server.calls.map((c) => c.path)).toEqual(["/api/login", "/api/me/bookings", "/api/refresh", "/api/me/bookings"]);
    expect(server.calls.at(-1)?.auth).toBe("Bearer tok2");
  });

  it("shares one refresh between requests that fail together", async () => {
    const auth = new Auth();
    await auth.login("0712345678", "matatu-2026");
    server.expire();
    const results = await Promise.all([auth.apiFetch("/api/me/bookings"), auth.apiFetch("/api/me/bookings")]);
    expect(results.map((r) => r.status)).toEqual([200, 200]);
    const refreshes = server.calls.filter((c) => c.path === "/api/refresh").length;
    expect(refreshes, "number of POST /api/refresh").toBe(1);
  });

  it("calls onExpired and throws when the refresh fails", async () => {
    const auth = new Auth();
    const expired = vi.fn();
    auth.onExpired = expired;
    await auth.login("0712345678", "matatu-2026");
    server.expire();
    server.killRefresh();
    await expect(auth.apiFetch("/api/me/bookings")).rejects.toBeInstanceOf(SessionExpiredError);
    expect(expired).toHaveBeenCalledTimes(1);
    expect(auth.loggedIn).toBe(false);
  });

  it("restores a session from the cookie after a reload", async () => {
    await new Auth().login("0712345678", "matatu-2026");
    const afterReload = new Auth(); // a new page: memory is empty, the cookie is not
    expect(await afterReload.restore()).toEqual({ phone: "0712345678", name: "Amina" });
    expect(afterReload.loggedIn).toBe(true);
  });

  it("logs out on the server and forgets the token", async () => {
    const auth = new Auth();
    await auth.login("0712345678", "matatu-2026");
    await auth.logout();
    expect(server.calls.at(-1)?.path).toBe("/api/logout");
    expect(auth.loggedIn).toBe(false);
    expect(await new Auth().restore(), "the cookie no longer works").toBeNull();
  });
});
