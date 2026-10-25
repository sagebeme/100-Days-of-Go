// What the Go server sends over the WebSocket, and how each message changes the seat map.

/** ServerMessage is one message from api/main.go. Seats are numbered from 1. */
export type ServerMessage =
  | { type: "snapshot"; version: number; taken: boolean[] }
  | { type: "seat"; version: number; seat: number; taken: boolean }
  | { type: "error"; message: string };

/** SeatState is what the page shows. version -1 means no snapshot has arrived yet. */
export interface SeatState {
  version: number;
  taken: boolean[];
  error: string;
}

export const EMPTY: SeatState = { version: -1, taken: [], error: "" };

/**
 * parseMessage turns the text of a WebSocket message into a ServerMessage, or null if it isn't
 * valid JSON or isn't one of the three types with the right field types.
 */
export function parseMessage(data: string): ServerMessage | null {
  // Your code here
  return null;
}

/**
 * applyMessage returns the new state (never changing the old one):
 * - snapshot: replaces version and taken, and clears error
 * - seat: sets that one seat, but only if a snapshot has arrived and msg.version is newer than
 *   state.version. An older or repeated update is ignored.
 * - error: sets error
 */
export function applyMessage(state: SeatState, msg: ServerMessage): SeatState {
  // Your code here
  return state;
}
