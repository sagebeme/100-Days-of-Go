// Day 70: the login flow. The access token lives only in this object's memory. The refresh token
// lives in an HttpOnly cookie that the server sets and JavaScript can never read.

export const API_URL: string = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

/** User is who is logged in. */
export interface User {
  phone: string;
  name: string;
}

/** AuthApi is everything the page needs from Auth. The App takes one, so tests can pass a fake. */
export interface AuthApi {
  restore(): Promise<User | null>;
  login(phone: string, password: string): Promise<User>;
  logout(): Promise<void>;
  apiFetch(path: string, init?: RequestInit): Promise<Response>;
  onExpired: () => void;
}

/** LoginError means the phone number or password was wrong. */
export class LoginError extends Error {
  constructor() {
    super("Wrong phone number or password");
    this.name = "LoginError";
  }
}

/** SessionExpiredError means the refresh token no longer works: the passenger must log in again. */
export class SessionExpiredError extends Error {
  constructor() {
    super("Your session has ended. Please log in again.");
    this.name = "SessionExpiredError";
  }
}

/**
 * Auth talks to the Go API. Every request uses credentials: "include" so the browser sends and
 * stores the refresh cookie across origins.
 * - login: POST /api/login {phone, password}. 200 gives {accessToken, user}; 401 throws LoginError.
 * - refresh: POST /api/refresh. 200 gives {accessToken, user}; anything else gives null and forgets
 *   the token. Calls made while a refresh is running share it, so the server sees one refresh.
 * - restore: refresh once on page load (memory is empty after a reload, but the cookie isn't).
 * - apiFetch: fetch API_URL + path with "Authorization: Bearer <token>". On 401, refresh and retry
 *   once. If the refresh fails, call onExpired and throw SessionExpiredError.
 * - logout: POST /api/logout and forget the token, even if the request fails.
 */
export class Auth implements AuthApi {
  #token: string | null = null; // # makes it private at runtime, not just in TypeScript
  #refreshing: Promise<User | null> | null = null;
  onExpired: () => void = () => {};

  get loggedIn(): boolean {
    return this.#token !== null;
  }

  async login(phone: string, password: string): Promise<User> {
    const res = await fetch(`${API_URL}/api/login`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ phone, password }),
    });
    if (res.status === 401) throw new LoginError();
    if (!res.ok) throw new Error(`Login failed (HTTP ${res.status})`);
    const body = (await res.json()) as { accessToken: string; user: User };
    this.#token = body.accessToken;
    return body.user;
  }

  refresh(): Promise<User | null> {
    // Single flight: if a refresh is already running, everyone waits for that one.
    this.#refreshing ??= this.#doRefresh().finally(() => {
      this.#refreshing = null;
    });
    return this.#refreshing;
  }

  async #doRefresh(): Promise<User | null> {
    try {
      const res = await fetch(`${API_URL}/api/refresh`, { method: "POST", credentials: "include" });
      if (!res.ok) {
        this.#token = null;
        return null;
      }
      const body = (await res.json()) as { accessToken: string; user: User };
      this.#token = body.accessToken;
      return body.user;
    } catch {
      this.#token = null;
      return null;
    }
  }

  restore(): Promise<User | null> {
    return this.refresh();
  }

  async apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
    const res = await this.#send(path, init);
    if (res.status !== 401) return res;
    const user = await this.refresh();
    if (user !== null) {
      const retry = await this.#send(path, init);
      if (retry.status !== 401) return retry;
    }
    this.#token = null;
    this.onExpired();
    throw new SessionExpiredError();
  }

  #send(path: string, init: RequestInit): Promise<Response> {
    const headers = new Headers(init.headers);
    if (this.#token !== null) headers.set("Authorization", `Bearer ${this.#token}`);
    return fetch(API_URL + path, { ...init, headers, credentials: "include" });
  }

  async logout(): Promise<void> {
    try {
      await fetch(`${API_URL}/api/logout`, { method: "POST", credentials: "include" });
    } catch {
      // offline: forget the token anyway; the cookie dies on its own
    } finally {
      this.#token = null;
    }
  }
}
