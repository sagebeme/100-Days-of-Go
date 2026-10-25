// Day 71: a seat map that updates live from Go.
import { useEffect, useRef, useState } from "react";
import { EMPTY, type SeatState, applyMessage } from "./seats";
import { type Connection, type Handlers, type Status, WS_URL, connectSeats } from "./socket";

const LABELS: Record<Status, string> = {
  connecting: "Connecting...",
  live: "Live",
  reconnecting: "Reconnecting...",
  closed: "Closed",
};

/**
 * LiveSeats connects on mount (with connect, so tests can pass a fake) and closes on unmount.
 * It shows <p role="status"> with LABELS[status], "N of M seats free", one button per seat
 * ("Seat 3: free" or "Seat 3: taken", disabled when taken) that sends {type: "hold", seat},
 * and the latest server error in role="alert".
 */
export function LiveSeats({
  connect = (h) => connectSeats(WS_URL, h),
}: {
  connect?: (handlers: Handlers) => Connection;
}) {
  const [seats, setSeats] = useState<SeatState>(EMPTY);
  const [status, setStatus] = useState<Status>("connecting");
  // A ref holds a value across renders without causing one. The click handler needs the connection.
  const conn = useRef<Connection | null>(null);

  useEffect(() => {
    conn.current = connect({
      // The function form, because messages arrive between renders.
      onMessage: (msg) => setSeats((s) => applyMessage(s, msg)),
      onStatus: setStatus,
    });
    return () => conn.current?.close();
    // Connect once, on mount. A new connect function on a later render must not reconnect.
  }, []);

  const free = seats.taken.filter((t) => !t).length;
  return (
    <main>
      <h1>Route 58 to Buruburu</h1>
      <p role="status">{LABELS[status]}</p>
      {seats.version >= 0 && (
        <p>
          {free} of {seats.taken.length} seats free
        </p>
      )}
      <div className="seats">
        {seats.taken.map((taken, i) => (
          // The index is the seat number here: seat 3 is always seat 3, so it's a stable key.
          <button key={i} type="button" disabled={taken} onClick={() => conn.current?.send({ type: "hold", seat: i + 1 })}>
            Seat {i + 1}: {taken ? "taken" : "free"}
          </button>
        ))}
      </div>
      {seats.error && <p role="alert">{seats.error}</p>}
    </main>
  );
}
