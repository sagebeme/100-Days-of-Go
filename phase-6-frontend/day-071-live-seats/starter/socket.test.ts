// @vitest-environment node
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { type SocketLike, type Status, backoff, connectSeats } from "./socket";

class FakeSocket implements SocketLike {
  onopen: ((ev: Event) => void) | null = null;
  onmessage: ((ev: MessageEvent) => void) | null = null;
  onclose: ((ev: CloseEvent) => void) | null = null;
  sent: string[] = [];
  closed = false;
  send(data: string) {
    this.sent.push(data);
  }
  close() {
    this.closed = true;
  }
  // Helpers for the test to play the server's part.
  open() {
    this.onopen?.(new Event("open"));
  }
  receive(data: string) {
    this.onmessage?.({ data } as MessageEvent);
  }
  drop() {
    this.onclose?.({} as CloseEvent);
  }
}

let sockets: FakeSocket[];
let statuses: Status[];
const make = () => {
  const s = new FakeSocket();
  sockets.push(s);
  return s;
};

beforeEach(() => {
  vi.useFakeTimers();
  sockets = [];
  statuses = [];
});
afterEach(() => {
  vi.useRealTimers();
});

describe("backoff", () => {
  it("doubles from 1 s and stops at 10 s", () => {
    expect([0, 1, 2, 3, 4, 5, 20].map(backoff)).toEqual([1000, 2000, 4000, 8000, 10000, 10000, 10000]);
  });
});

describe("connectSeats", () => {
  it("reports connecting, then live, and passes on valid messages only", () => {
    const got: unknown[] = [];
    connectSeats("ws://x", { onMessage: (m) => got.push(m), onStatus: (s) => statuses.push(s) }, make);
    expect(statuses).toEqual(["connecting"]);
    sockets[0].open();
    expect(statuses).toEqual(["connecting", "live"]);
    sockets[0].receive('{"type":"error","message":"hi"}');
    sockets[0].receive("garbage");
    expect(got).toEqual([{ type: "error", message: "hi" }]);
  });

  it("sends JSON only while open", () => {
    const conn = connectSeats("ws://x", { onMessage: () => {}, onStatus: () => {} }, make);
    expect(conn.send({ type: "hold", seat: 3 }), "send before open").toBe(false);
    sockets[0].open();
    expect(conn.send({ type: "hold", seat: 3 })).toBe(true);
    expect(sockets[0].sent).toEqual(['{"type":"hold","seat":3}']);
  });

  it("reconnects after a drop, waiting longer each time", () => {
    connectSeats("ws://x", { onMessage: () => {}, onStatus: (s) => statuses.push(s) }, make);
    sockets[0].drop();
    expect(statuses.at(-1)).toBe("reconnecting");
    vi.advanceTimersByTime(999);
    expect(sockets.length, "no new socket before 1 s").toBe(1);
    vi.advanceTimersByTime(1);
    expect(sockets.length, "a new socket at 1 s").toBe(2);
    sockets[1].drop();
    vi.advanceTimersByTime(1999);
    expect(sockets.length, "the second retry waits 2 s").toBe(2);
    vi.advanceTimersByTime(1);
    expect(sockets.length).toBe(3);
  });

  it("starts the wait again from 1 s after a successful connect", () => {
    connectSeats("ws://x", { onMessage: () => {}, onStatus: (s) => statuses.push(s) }, make);
    sockets[0].drop();
    vi.advanceTimersByTime(1000);
    sockets[1].drop();
    vi.advanceTimersByTime(2000);
    sockets[2].open();
    expect(statuses.at(-1)).toBe("live");
    sockets[2].drop();
    vi.advanceTimersByTime(1000);
    expect(sockets.length).toBe(4);
  });

  it("stops for good on close()", () => {
    const conn = connectSeats("ws://x", { onMessage: () => {}, onStatus: (s) => statuses.push(s) }, make);
    sockets[0].open();
    sockets[0].drop();
    conn.close();
    vi.advanceTimersByTime(60_000);
    expect(sockets.length, "no reconnect after close").toBe(1);
    expect(statuses.at(-1)).toBe("closed");
  });

  it("closes the socket on close()", () => {
    const conn = connectSeats("ws://x", { onMessage: () => {}, onStatus: () => {} }, make);
    sockets[0].open();
    conn.close();
    expect(sockets[0].closed).toBe(true);
  });
});
