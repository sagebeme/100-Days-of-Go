// @vitest-environment node
import { describe, expect, it } from "vitest";
import { createFakeFetch, json } from "./fakeFetch";

describe("json", () => {
  it("makes a JSON response", async () => {
    const res = json(201, { id: "T1" });
    expect(res.status).toBe(201);
    expect(res.headers.get("Content-Type")).toBe("application/json");
    expect(await res.json()).toEqual({ id: "T1" });
  });
});

describe("createFakeFetch", () => {
  it("routes by method and path, ignoring the origin and query", async () => {
    const fake = createFakeFetch({
      "GET /api/trips": () => json(200, ["trips"]),
      "POST /api/bookings": async (req) => json(201, await req.json()),
    });
    const a = await fake.fetch("http://localhost:8080/api/trips?route=58");
    expect(await a.json()).toEqual(["trips"]);
    const b = await fake.fetch(new URL("http://x/api/bookings"), { method: "post", body: JSON.stringify({ seat: 3 }) });
    expect(b.status).toBe(201);
    expect(await b.json(), "the handler gets the request body").toEqual({ seat: 3 });
    expect(fake.calls).toEqual(["GET /api/trips", "POST /api/bookings"]);
  });

  it("answers 404 with a helpful message when there is no fake", async () => {
    const fake = createFakeFetch({});
    const res = await fake.fetch("http://localhost:8080/api/nope", { method: "DELETE" });
    expect(res.status).toBe(404);
    expect(await res.json()).toEqual({ error: "no fake for DELETE /api/nope" });
    expect(fake.calls).toEqual(["DELETE /api/nope"]);
  });

  it("rejects when a handler throws, like a network failure", async () => {
    const fake = createFakeFetch({
      "GET /api/trips": () => {
        throw new TypeError("Failed to fetch");
      },
    });
    await expect(fake.fetch("http://localhost:8080/api/trips")).rejects.toThrow("Failed to fetch");
  });

  it("accepts an abort signal without complaint", async () => {
    const fake = createFakeFetch({ "GET /api/trips": () => json(200, []) });
    const res = await fake.fetch("http://localhost:8080/api/trips", { signal: new AbortController().signal });
    expect(res.status).toBe(200);
  });
});
