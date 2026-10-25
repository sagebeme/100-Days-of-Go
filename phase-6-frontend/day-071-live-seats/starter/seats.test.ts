// @vitest-environment node
import { describe, expect, it } from "vitest";
import { EMPTY, type SeatState, applyMessage, parseMessage } from "./seats";

describe("parseMessage", () => {
  const cases: { data: string; want: unknown }[] = [
    { data: '{"type":"snapshot","version":7,"taken":[false,true]}', want: { type: "snapshot", version: 7, taken: [false, true] } },
    { data: '{"type":"seat","version":8,"seat":3,"taken":true}', want: { type: "seat", version: 8, seat: 3, taken: true } },
    { data: '{"type":"error","message":"Seat 3 is already taken"}', want: { type: "error", message: "Seat 3 is already taken" } },
    { data: "not json", want: null },
    { data: "null", want: null },
    { data: '{"type":"seat","version":8,"seat":"3","taken":true}', want: null },
    { data: '{"type":"snapshot","version":1,"taken":[1,0]}', want: null },
    { data: '{"type":"party"}', want: null },
  ];
  for (const c of cases) {
    it(c.data, () => {
      expect(parseMessage(c.data)).toEqual(c.want);
    });
  }
});

describe("applyMessage", () => {
  const snap: SeatState = { version: 7, taken: [false, true, false], error: "" };

  it("takes a snapshot whole", () => {
    expect(applyMessage(EMPTY, { type: "snapshot", version: 7, taken: [false, true, false] })).toEqual(snap);
  });
  it("applies a newer seat update without changing the old state", () => {
    const got = applyMessage(snap, { type: "seat", version: 8, seat: 3, taken: true });
    expect(got).toEqual({ version: 8, taken: [false, true, true], error: "" });
    expect(snap.taken, "the old state is untouched").toEqual([false, true, false]);
  });
  it("ignores an update older than what it has", () => {
    expect(applyMessage(snap, { type: "seat", version: 6, seat: 1, taken: true })).toEqual(snap);
  });
  it("ignores a repeated update", () => {
    expect(applyMessage(snap, { type: "seat", version: 7, seat: 1, taken: true })).toEqual(snap);
  });
  it("ignores updates before the first snapshot", () => {
    expect(applyMessage(EMPTY, { type: "seat", version: 3, seat: 1, taken: true })).toEqual(EMPTY);
  });
  it("records an error, and a snapshot clears it", () => {
    const withError = applyMessage(snap, { type: "error", message: "Seat 2 is already taken" });
    expect(withError.error).toBe("Seat 2 is already taken");
    expect(applyMessage(withError, { type: "snapshot", version: 9, taken: [true] }).error).toBe("");
  });
});
