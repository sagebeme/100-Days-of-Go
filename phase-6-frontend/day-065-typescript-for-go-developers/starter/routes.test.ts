// @vitest-environment node
// (No browser needed today, so the tests skip the fake DOM and start faster.)
import { describe, expect, it } from "vitest";
import {
  type Payment,
  type Route,
  ROUTES,
  describePayment,
  fareFor,
  groupBy,
  parseRoute,
  totalFare,
} from "./routes";

describe("describePayment", () => {
  const cases: { name: string; payment: Payment; want: string }[] = [
    {
      name: "mpesa",
      payment: { kind: "mpesa", phone: "0712345678", amountKsh: 100 },
      want: "Ksh100 by M-Pesa from 0712345678",
    },
    { name: "cash", payment: { kind: "cash", amountKsh: 150 }, want: "Ksh150 in cash" },
    {
      name: "card",
      payment: { kind: "card", last4: "4242", amountKsh: 100 },
      want: "Ksh100 by card ending 4242",
    },
  ];
  for (const c of cases) {
    it(c.name, () => {
      const got = describePayment(c.payment);
      expect(got, `describePayment(${JSON.stringify(c.payment)})`).toBe(c.want);
    });
  }
});

describe("fareFor", () => {
  it("finds a route's fare", () => {
    expect(fareFor(ROUTES, "58")).toBe(150);
    expect(fareFor(ROUTES, "125")).toBe(100);
  });
  it("gives undefined for a route that isn't there", () => {
    expect(fareFor(ROUTES, "999")).toBeUndefined();
  });
});

describe("groupBy", () => {
  it("groups routes by where they board", () => {
    const got = groupBy(ROUTES, (r) => r.boardsAt);
    expect([...got.keys()]).toEqual(["Kencom / Ambassadeur", "Railways", "Ronald Ngala Street", "Ambassadeur"]);
    expect(got.get("Railways")?.map((r) => r.number)).toEqual(["111", "125"]);
  });
  it("works for any type, not just routes", () => {
    const got = groupBy(["boda", "bus", "matatu", "mkokoteni"], (s) => s[0]);
    expect(Object.fromEntries(got)).toEqual({ b: ["boda", "bus"], m: ["matatu", "mkokoteni"] });
  });
  it("gives an empty map for an empty list", () => {
    expect(groupBy([] as Route[], (r) => r.number).size).toBe(0);
  });
});

describe("parseRoute", () => {
  it("accepts a well-formed route", () => {
    const raw: unknown = JSON.parse(
      '{"number":"46K","operator":"Citi Hoppa","boardsAt":"Kencom / Ambassadeur","goesTo":"Kawangware","fareKsh":100}',
    );
    expect(parseRoute(raw)).toEqual(ROUTES[0]);
  });
  const bad: { name: string; json: string }[] = [
    { name: "not an object", json: '"46K"' },
    { name: "null", json: "null" },
    { name: "missing fare", json: '{"number":"46K","operator":"x","boardsAt":"y","goesTo":"z"}' },
    { name: "fare is a string", json: '{"number":"46K","operator":"x","boardsAt":"y","goesTo":"z","fareKsh":"100"}' },
    { name: "number is a number", json: '{"number":46,"operator":"x","boardsAt":"y","goesTo":"z","fareKsh":100}' },
  ];
  for (const c of bad) {
    it(`rejects ${c.name}`, () => {
      expect(() => parseRoute(JSON.parse(c.json))).toThrow(/invalid route/);
    });
  }
});

describe("totalFare", () => {
  it("adds the fares it looks up", async () => {
    const lookup = async (n: string) => fareFor(ROUTES, n) ?? 0;
    await expect(totalFare(["46K", "58", "125"], lookup)).resolves.toBe(350);
  });
  it("starts every lookup before waiting for any of them", async () => {
    const started: string[] = [];
    const lookup = (n: string) => {
      started.push(n);
      return new Promise<number>((resolve) => setTimeout(() => resolve(100), 0));
    };
    const total = totalFare(["111", "125", "45"], lookup);
    expect(started, "lookups started before the first one finished").toEqual(["111", "125", "45"]);
    await expect(total).resolves.toBe(300);
  });
  it("fails if any lookup fails", async () => {
    const lookup = async (n: string) => {
      if (n === "999") throw new Error("no such route");
      return 100;
    };
    await expect(totalFare(["46K", "999"], lookup)).rejects.toThrow("no such route");
  });
});
