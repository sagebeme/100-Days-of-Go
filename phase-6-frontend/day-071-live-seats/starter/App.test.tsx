import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { LiveSeats } from "./App";
import type { Connection, Handlers } from "./socket";

/** fakeConnect stands in for connectSeats. The test drives the handlers by hand. */
function fakeConnect() {
  let handlers: Handlers = { onMessage: () => {}, onStatus: () => {} };
  const conn: Connection = { send: vi.fn(() => true), close: vi.fn() };
  const connect = vi.fn((h: Handlers) => {
    handlers = h;
    h.onStatus("connecting");
    return conn;
  });
  return { connect, conn, server: () => handlers };
}

function seat(n: number): HTMLButtonElement {
  return screen.getByRole("button", { name: new RegExp(`^Seat ${n}:`) }) as HTMLButtonElement;
}

describe("LiveSeats", () => {
  it("shows the connection status", () => {
    const f = fakeConnect();
    render(<LiveSeats connect={f.connect} />);
    expect(screen.getByRole("status").textContent).toBe("Connecting...");
    act(() => f.server().onStatus("live"));
    expect(screen.getByRole("status").textContent).toBe("Live");
    act(() => f.server().onStatus("reconnecting"));
    expect(screen.getByRole("status").textContent).toBe("Reconnecting...");
  });

  it("draws the snapshot and live changes", () => {
    const f = fakeConnect();
    render(<LiveSeats connect={f.connect} />);
    act(() => f.server().onMessage({ type: "snapshot", version: 1, taken: [false, true, false, false] }));
    expect(seat(1).textContent).toBe("Seat 1: free");
    expect(seat(2).textContent).toBe("Seat 2: taken");
    expect(seat(2).disabled).toBe(true);
    screen.getByText("3 of 4 seats free");
    act(() => f.server().onMessage({ type: "seat", version: 2, seat: 1, taken: true }));
    expect(seat(1).textContent).toBe("Seat 1: taken");
    screen.getByText("2 of 4 seats free");
  });

  it("asks the server to hold a free seat", async () => {
    const f = fakeConnect();
    render(<LiveSeats connect={f.connect} />);
    act(() => f.server().onMessage({ type: "snapshot", version: 1, taken: [false, false] }));
    await userEvent.setup().click(seat(2));
    expect(f.conn.send).toHaveBeenCalledWith({ type: "hold", seat: 2 });
  });

  it("shows the server's error", () => {
    const f = fakeConnect();
    render(<LiveSeats connect={f.connect} />);
    act(() => f.server().onMessage({ type: "error", message: "Seat 2 is already taken" }));
    expect(screen.getByRole("alert").textContent).toBe("Seat 2 is already taken");
  });

  it("closes the connection when it goes away", () => {
    const f = fakeConnect();
    const { unmount } = render(<LiveSeats connect={f.connect} />);
    unmount();
    expect(f.conn.close).toHaveBeenCalled();
  });
});
