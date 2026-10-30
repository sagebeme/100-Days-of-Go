import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, LoginError, SessionExpiredError, createClient } from "./api";
import { type Handler, createFakeFetch, json } from "./fakeFetch";

const BASE = "http://api.test";
const USER = { phone: "0712345678", name: "Amina" };
const TRIP = { id: "T1", route: "58", goesTo: "Buruburu", boardsAt: "Ambassadeur", departs: "07:45", fareKsh: 150, seatsLeft: 3 };
const INPUT = { tripId: "T1", seat: 3, name: "Amina", phone: "0712345678" };

/** server fakes the Go API's auth: login gives tok1, each refresh the next token; only the newest works. */
function server(extra: Record<string, Handler> = {}) {
  let n = 0;
  let current = "";
  let cookie = false;
  const credentials: RequestCredentials[] = [];
  const issue = () => {
    current = `tok${++n}`;
    cookie = true;
    return json(200, { accessToken: current, user: USER });
  };
  const wrap = (h: Handler): Handler => (req) => {
    credentials.push(req.credentials);
    return h(req);
  };
  const routes: Record<string, Handler> = {
    "POST /api/login": async (req) => ((await req.json()).password === "matatu-2026" ? issue() : json(401, { error: "wrong" })),
    "POST /api/refresh": () => (cookie ? issue() : json(401, { error: "please log in" })),
    "POST /api/logout": () => {
      cookie = false;
      return new Response(null, { status: 204 });
    },
    ...extra,
  };
  const fake = createFakeFetch(Object.fromEntries(Object.entries(routes).map(([k, h]) => [k, wrap(h)])));
  vi.stubGlobal("fetch", fake.fetch);
  return {
    fake,
    credentials,
    authorised: (req: Request) => current !== "" && req.headers.get("Authorization") === `Bearer ${current}`,
    expire: () => {
      current = "gone";
    },
    dropCookie: () => {
      cookie = false;
    },
  };
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("reading trips and seats", () => {
  it("lists trips", async () => {
    server({ "GET /api/trips": () => json(200, [TRIP]) });
    expect(await createClient(BASE).listTrips()).toEqual([TRIP]);
  });
  it("gets a trip's seats", async () => {
    const s = server({ "GET /api/trips/T1/seats": () => json(200, { tripId: "T1", seatCount: 14, taken: [2, 5] }) });
    expect(await createClient(BASE).getSeats("T1")).toEqual({ tripId: "T1", seatCount: 14, taken: [2, 5] });
    expect(s.fake.calls).toEqual(["GET /api/trips/T1/seats"]);
  });
  it("throws an ApiError with the server's message", async () => {
    server({ "GET /api/trips": () => json(503, { error: "database is down" }) });
    const err = await createClient(BASE).listTrips().catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).status).toBe(503);
    expect((err as ApiError).message).toBe("database is down");
  });
  it("throws for a missing trip", async () => {
    server({ "GET /api/trips/T9/seats": () => json(404, { error: "no such trip" }) });
    await expect(createClient(BASE).getSeats("T9")).rejects.toThrow("no such trip");
  });
});

describe("logging in", () => {
  it("logs in, restores from the cookie, and logs out", async () => {
    const s = server();
    expect(await createClient(BASE).login("0712345678", "matatu-2026")).toEqual(USER);
    expect(await createClient(BASE).restore(), "a fresh client (a reload) restores from the cookie").toEqual(USER);
    await createClient(BASE).logout();
    expect(await createClient(BASE).restore(), "after logout the cookie is gone").toBeNull();
    expect(s.credentials.every((c) => c === "include"), "every request sends credentials").toBe(true);
  });
  it("throws LoginError for a wrong password", async () => {
    server();
    await expect(createClient(BASE).login("0712345678", "1234")).rejects.toBeInstanceOf(LoginError);
  });
});

describe("booking", () => {
  function bookingServer(answer: () => Response) {
    const s = server({
      "POST /api/bookings": (req) => (s.authorised(req) ? answer() : json(401, { error: "please log in" })),
    });
    return s;
  }
  const booked = { reference: "K7Q2M9XA", tripId: "T1", seat: 3, phone: "254712345678", totalKsh: 150 };

  const cases: { name: string; answer: () => Response; want: unknown }[] = [
    { name: "201 is booked", answer: () => json(201, booked), want: { kind: "booked", booking: booked } },
    {
      name: "422 is invalid",
      answer: () => json(422, { errors: { phone: "Enter a Kenyan mobile number like 0712 345 678" } }),
      want: { kind: "invalid", errors: { phone: "Enter a Kenyan mobile number like 0712 345 678" } },
    },
    { name: "409 is a conflict", answer: () => json(409, { error: "Seat 3 was just taken" }), want: { kind: "conflict", message: "Seat 3 was just taken" } },
  ];
  for (const c of cases) {
    it(c.name, async () => {
      bookingServer(c.answer);
      const client = createClient(BASE);
      await client.login("0712345678", "matatu-2026");
      expect(await client.book(INPUT)).toEqual(c.want);
    });
  }

  it("throws an ApiError for anything else", async () => {
    bookingServer(() => json(500, { error: "boom" }));
    const client = createClient(BASE);
    await client.login("0712345678", "matatu-2026");
    await expect(client.book(INPUT)).rejects.toBeInstanceOf(ApiError);
  });

  it("refreshes an expired token once and retries", async () => {
    const s = bookingServer(() => json(201, booked));
    const client = createClient(BASE);
    await client.login("0712345678", "matatu-2026");
    s.expire();
    expect((await client.book(INPUT)).kind).toBe("booked");
    expect(s.fake.calls).toEqual(["POST /api/login", "POST /api/bookings", "POST /api/refresh", "POST /api/bookings"]);
  });

  it("calls onExpired and throws when the refresh fails", async () => {
    const s = bookingServer(() => json(201, booked));
    const client = createClient(BASE);
    const expired = vi.fn();
    client.onExpired = expired;
    await client.login("0712345678", "matatu-2026");
    s.expire();
    s.dropCookie();
    await expect(client.book(INPUT)).rejects.toBeInstanceOf(SessionExpiredError);
    expect(expired).toHaveBeenCalledTimes(1);
  });
});
