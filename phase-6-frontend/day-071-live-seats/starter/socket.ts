// A WebSocket that reconnects by itself. The browser's WebSocket never retries, so we do.
import { type ServerMessage, parseMessage } from "./seats";

export const WS_URL: string = import.meta.env.VITE_WS_URL ?? "ws://localhost:8080/ws/seats";

/** Status is what the page tells the passenger about the connection. */
export type Status = "connecting" | "live" | "reconnecting" | "closed";

/** SocketLike is the part of the browser's WebSocket we use, so tests can pass a fake. */
export interface SocketLike {
  onopen: ((ev: Event) => void) | null;
  onmessage: ((ev: MessageEvent) => void) | null;
  onclose: ((ev: CloseEvent) => void) | null;
  send(data: string): void;
  close(): void;
}

export interface Handlers {
  onMessage: (msg: ServerMessage) => void;
  onStatus: (status: Status) => void;
}

export interface Connection {
  /** send sends msg as JSON if the socket is open, and reports whether it did. */
  send(msg: object): boolean;
  /** close closes the socket and stops reconnecting. */
  close(): void;
}

/** backoff is how long to wait before reconnect attempt number attempt (from 0): 1 s, 2 s, 4 s, 8 s, then 10 s. */
export function backoff(attempt: number): number {
  // Your code here
  return 0;
}

/**
 * connectSeats opens a socket with makeSocket(url) and reports "connecting". On open: "live" and
 * the attempt count goes back to 0. Each valid message goes to onMessage (invalid ones are
 * ignored). On close: "reconnecting", then a new socket after backoff(attempt), with attempt going
 * up each time. After close(), it reports "closed" and never reconnects.
 */
export function connectSeats(
  url: string,
  handlers: Handlers,
  makeSocket: (url: string) => SocketLike = (u) => new WebSocket(u),
): Connection {
  // Your code here
  return { send: () => false, close: () => {} };
}
