// The component under test. It's finished: today you build the tools that test it.
import { useEffect, useState } from "react";

export const API_URL: string = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export interface Trip {
  id: string;
  route: string;
  goesTo: string;
  departsAt: string; // ISO 8601, e.g. "2026-10-30T07:45:00Z"
  seatsLeft: number;
  fareKsh: number;
}

/** describeTrip is "58 to Buruburu, leaves in 12 min, 3 seats left, Ksh150", or "..., full". */
export function describeTrip(t: Trip, now: number): string {
  const minutes = Math.max(0, Math.round((Date.parse(t.departsAt) - now) / 60_000));
  const when = minutes === 0 ? "leaving now" : `leaves in ${minutes} min`;
  const seats = t.seatsLeft === 0 ? "full" : `${t.seatsLeft} ${t.seatsLeft === 1 ? "seat" : "seats"} left, Ksh${t.fareKsh}`;
  return `${t.route} to ${t.goesTo}, ${when}, ${seats}`;
}

type State = { kind: "loading" } | { kind: "error" } | { kind: "done"; trips: Trip[] };

/** TripList loads GET /api/trips and lists them, with a filter for trips that still have seats. */
export function TripList() {
  const [state, setState] = useState<State>({ kind: "loading" });
  const [attempt, setAttempt] = useState(0);
  const [onlyFree, setOnlyFree] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setState({ kind: "loading" });
    fetch(`${API_URL}/api/trips`)
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json() as Promise<Trip[]>;
      })
      .then((trips) => !cancelled && setState({ kind: "done", trips }))
      .catch(() => !cancelled && setState({ kind: "error" }));
    return () => {
      cancelled = true;
    };
  }, [attempt]);

  const now = Date.now();
  return (
    <main>
      <h1>Next trips</h1>
      <label>
        <input type="checkbox" checked={onlyFree} onChange={(e) => setOnlyFree(e.target.checked)} /> Only trips with seats
      </label>
      {state.kind === "loading" && <p role="status">Loading trips...</p>}
      {state.kind === "error" && (
        <p role="alert">
          Couldn't load trips.{" "}
          <button type="button" onClick={() => setAttempt((a) => a + 1)}>
            Try again
          </button>
        </p>
      )}
      {state.kind === "done" && (
        <ul aria-label="Trips">
          {state.trips
            .filter((t) => !onlyFree || t.seatsLeft > 0)
            .map((t) => (
              <li key={t.id}>{describeTrip(t, now)}</li>
            ))}
        </ul>
      )}
    </main>
  );
}
