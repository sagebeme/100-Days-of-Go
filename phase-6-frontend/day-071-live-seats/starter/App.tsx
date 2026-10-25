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
  // Your code here
  return <main></main>;
}
