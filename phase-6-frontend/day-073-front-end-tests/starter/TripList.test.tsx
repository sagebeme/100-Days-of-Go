// Tests for TripList, using your fake fetch. Read them as examples: arrange, act, assert.
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createFakeFetch, json } from "./fakeFetch";
import { type Trip, TripList, describeTrip } from "./TripList";

const NOW = Date.parse("2026-10-30T07:30:00Z");
const TRIPS: Trip[] = [
  { id: "T1", route: "58", goesTo: "Buruburu", departsAt: "2026-10-30T07:42:00Z", seatsLeft: 3, fareKsh: 150 },
  { id: "T2", route: "111", goesTo: "Ngong", departsAt: "2026-10-30T07:35:00Z", seatsLeft: 0, fareKsh: 100 },
  { id: "T3", route: "125", goesTo: "Ongata Rongai", departsAt: "2026-10-30T07:30:20Z", seatsLeft: 1, fareKsh: 100 },
];

beforeEach(() => {
  // Fake only the clock, so "leaves in 12 min" is the same on every run. Timers stay real.
  vi.useFakeTimers({ toFake: ["Date"] });
  vi.setSystemTime(NOW);
});
afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe("describeTrip", () => {
  const cases: { name: string; trip: Trip; want: string }[] = [
    { name: "seats left", trip: TRIPS[0], want: "58 to Buruburu, leaves in 12 min, 3 seats left, Ksh150" },
    { name: "full", trip: TRIPS[1], want: "111 to Ngong, leaves in 5 min, full" },
    { name: "one seat, leaving now", trip: TRIPS[2], want: "125 to Ongata Rongai, leaving now, 1 seat left, Ksh100" },
  ];
  for (const c of cases) {
    it(c.name, () => {
      expect(describeTrip(c.trip, NOW)).toBe(c.want);
    });
  }
});

describe("TripList", () => {
  it("loads and lists the trips", async () => {
    const fake = createFakeFetch({ "GET /api/trips": () => json(200, TRIPS) });
    vi.stubGlobal("fetch", fake.fetch);
    render(<TripList />);
    screen.getByRole("status");
    const items = within(await screen.findByRole("list", { name: "Trips" })).getAllByRole("listitem");
    expect(items.map((li) => li.textContent)).toEqual(TRIPS.map((t) => describeTrip(t, NOW)));
    expect(fake.calls).toEqual(["GET /api/trips"]);
  });

  it("hides full trips when asked", async () => {
    vi.stubGlobal("fetch", createFakeFetch({ "GET /api/trips": () => json(200, TRIPS) }).fetch);
    render(<TripList />);
    await screen.findByRole("list", { name: "Trips" });
    await userEvent.setup().click(screen.getByRole("checkbox", { name: "Only trips with seats" }));
    expect(screen.getAllByRole("listitem").map((li) => li.textContent)).not.toContain(describeTrip(TRIPS[1], NOW));
    expect(screen.getAllByRole("listitem")).toHaveLength(2);
  });

  it("shows an error and recovers on Try again", async () => {
    let fail = true;
    const fake = createFakeFetch({ "GET /api/trips": () => (fail ? json(503, { error: "down" }) : json(200, TRIPS)) });
    vi.stubGlobal("fetch", fake.fetch);
    render(<TripList />);
    const alert = await screen.findByRole("alert");
    fail = false;
    await userEvent.setup().click(within(alert).getByRole("button", { name: "Try again" }));
    await screen.findByRole("list", { name: "Trips" });
    expect(fake.calls).toEqual(["GET /api/trips", "GET /api/trips"]);
  });

  it("treats a network failure like any other error", async () => {
    vi.stubGlobal(
      "fetch",
      createFakeFetch({
        "GET /api/trips": () => {
          throw new TypeError("Failed to fetch");
        },
      }).fetch,
    );
    render(<TripList />);
    expect((await screen.findByRole("alert")).textContent).toContain("Couldn't load trips.");
  });
});
