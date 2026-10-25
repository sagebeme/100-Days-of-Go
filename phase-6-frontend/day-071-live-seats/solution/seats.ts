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
  let v: unknown;
  try {
    v = JSON.parse(data);
  } catch {
    return null;
  }
  if (typeof v !== "object" || v === null) return null;
  const m = v as Record<string, unknown>;
  switch (m.type) {
    case "snapshot":
      if (typeof m.version === "number" && Array.isArray(m.taken) && m.taken.every((t) => typeof t === "boolean")) {
        return { type: "snapshot", version: m.version, taken: m.taken as boolean[] };
      }
      return null;
    case "seat":
      if (typeof m.version === "number" && typeof m.seat === "number" && typeof m.taken === "boolean") {
        return { type: "seat", version: m.version, seat: m.seat, taken: m.taken };
      }
      return null;
    case "error":
      return typeof m.message === "string" ? { type: "error", message: m.message } : null;
    default:
      return null;
  }
}

/**
 * applyMessage returns the new state (never changing the old one):
 * - snapshot: replaces version and taken, and clears error
 * - seat: sets that one seat, but only if a snapshot has arrived and msg.version is newer than
 *   state.version. An older or repeated update is ignored.
 * - error: sets error
 */
export function applyMessage(state: SeatState, msg: ServerMessage): SeatState {
  switch (msg.type) {
    case "snapshot":
      return { version: msg.version, taken: [...msg.taken], error: "" };
    case "seat": {
      if (state.version < 0 || msg.version <= state.version) return state;
      const taken = [...state.taken];
      taken[msg.seat - 1] = msg.taken;
      return { ...state, version: msg.version, taken };
    }
    case "error":
      return { ...state, error: msg.message };
  }
}
