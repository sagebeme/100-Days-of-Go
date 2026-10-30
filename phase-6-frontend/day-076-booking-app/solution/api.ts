// The client for the capstone's Go API (see api/main.go for every endpoint). The types and errors
// are the contract the tests use; keep them as they are.

export const API_URL: string = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export interface User {
  phone: string;
  name: string;
}

export interface Trip {
  id: string;
  route: string;
  goesTo: string;
  boardsAt: string;
  departs: string; // "07:45"
  fareKsh: number;
  seatsLeft: number;
}

export interface SeatInfo {
  tripId: string;
  seatCount: number;
  taken: number[];
}

export interface BookingInput {
  tripId: string;
  seat: number;
  name: string;
  phone: string;
}

export interface Booking {
  reference: string;
  tripId: string;
  seat: number;
  phone: string;
  totalKsh: number;
}

export type Field = "name" | "phone" | "seat";
export type FieldErrors = Partial<Record<Field, string>>;

/** BookingResult is what POST /api/bookings can mean: 201, 422 or 409. */
export type BookingResult =
  | { kind: "booked"; booking: Booking }
  | { kind: "invalid"; errors: FieldErrors }
  | { kind: "conflict"; message: string };

/** ApiError is any other non-2xx answer. message is the body's "error", or "HTTP STATUS". */
export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export class LoginError extends Error {
  constructor() {
    super("Wrong phone number or password");
    this.name = "LoginError";
  }
}

export class SessionExpiredError extends Error {
  constructor() {
    super("Your session has ended. Please log in again.");
    this.name = "SessionExpiredError";
  }
}

/** Client is everything the App needs. Tests pass a fake one. */
export interface Client {
  restore(): Promise<User | null>;
  login(phone: string, password: string): Promise<User>;
  logout(): Promise<void>;
  listTrips(): Promise<Trip[]>;
  getSeats(tripId: string): Promise<SeatInfo>;
  book(input: BookingInput): Promise<BookingResult>;
  /** onExpired is called when a refresh fails during book, just before it throws SessionExpiredError. */
  onExpired: () => void;
}

async function errorFrom(res: Response): Promise<ApiError> {
  let message = `HTTP ${res.status}`;
  try {
    const body = (await res.json()) as { error?: unknown };
    if (typeof body.error === "string") message = body.error;
  } catch {
    // not JSON
  }
  return new ApiError(res.status, message);
}

/**
 * createClient talks to the Go API at baseUrl. Every request uses credentials: "include".
 * The requirements are in the README; the tests in api.test.ts spell out each one.
 */
export function createClient(baseUrl: string = API_URL): Client {
  let token: string | null = null; // memory only (Day 70)
  let refreshing: Promise<User | null> | null = null;

  function send(path: string, init: RequestInit = {}): Promise<Response> {
    const headers = new Headers(init.headers);
    if (token !== null) headers.set("Authorization", `Bearer ${token}`);
    return fetch(baseUrl + path, { ...init, headers, credentials: "include" });
  }

  async function getJSON<T>(path: string): Promise<T> {
    const res = await send(path);
    if (!res.ok) throw await errorFrom(res);
    return (await res.json()) as T;
  }

  async function takeTokens(res: Response): Promise<User> {
    const body = (await res.json()) as { accessToken: string; user: User };
    token = body.accessToken;
    return body.user;
  }

  function refresh(): Promise<User | null> {
    refreshing ??= (async () => {
      try {
        const res = await send("/api/refresh", { method: "POST" });
        if (res.ok) return await takeTokens(res);
      } catch {
        // offline: same as a failed refresh
      }
      token = null;
      return null;
    })().finally(() => {
      refreshing = null;
    });
    return refreshing;
  }

  const client: Client = {
    onExpired: () => {},
    restore: refresh,
    async login(phone, password) {
      const res = await send("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ phone, password }),
      });
      if (res.status === 401) throw new LoginError();
      if (!res.ok) throw await errorFrom(res);
      return takeTokens(res);
    },
    async logout() {
      try {
        await send("/api/logout", { method: "POST" });
      } catch {
        // offline: forget the token anyway
      } finally {
        token = null;
      }
    },
    listTrips: () => getJSON<Trip[]>("/api/trips"),
    getSeats: (tripId) => getJSON<SeatInfo>(`/api/trips/${encodeURIComponent(tripId)}/seats`),
    async book(input) {
      const post = () =>
        send("/api/bookings", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(input),
        });
      let res = await post();
      if (res.status === 401) {
        const user = await refresh();
        if (user !== null) res = await post();
        if (user === null || res.status === 401) {
          token = null;
          client.onExpired();
          throw new SessionExpiredError();
        }
      }
      if (res.status === 201) return { kind: "booked", booking: (await res.json()) as Booking };
      if (res.status === 422) return { kind: "invalid", errors: ((await res.json()) as { errors: FieldErrors }).errors };
      if (res.status === 409) return { kind: "conflict", message: (await errorFrom(res)).message };
      throw await errorFrom(res);
    },
  };
  return client;
}
