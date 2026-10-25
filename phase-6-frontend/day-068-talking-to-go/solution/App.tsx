// Day 68: loading, error and success states for a request to Go.
import { useEffect, useState } from "react";
import { type Route, fetchRoutes } from "./api";

/** Status is the request's state. Only one of these is true at a time, so it's a union, not three booleans. */
type Status =
  | { kind: "loading" }
  | { kind: "error"; message: string }
  | { kind: "done"; routes: Route[] };

function label(r: Route): string {
  const title = r.number ? `${r.number} to ${r.goesTo}` : `To ${r.goesTo}`;
  return r.fareKsh === undefined ? `${title}, ask the conductor` : `${title}, Ksh${r.fareKsh}`;
}

/**
 * RouteSearch has an input labelled "Destination". Whenever the text changes (and once on load)
 * it calls fetchRoutes, cancelling the previous request with an AbortController. It shows:
 * - while loading: <p role="status">Loading routes...</p>
 * - on error: <div role="alert">Couldn't load routes: MESSAGE <button>Try again</button></div>
 * - no results: "No routes match."
 * - results: <ul aria-label="Routes"> with "111 to Ngong, Ksh100" or "To Kibera, ask the conductor"
 */
export function RouteSearch() {
  const [query, setQuery] = useState("");
  const [attempt, setAttempt] = useState(0); // bumping it re-runs the effect: that's Try again
  const [status, setStatus] = useState<Status>({ kind: "loading" });

  useEffect(() => {
    const controller = new AbortController();
    setStatus({ kind: "loading" });
    fetchRoutes(query, controller.signal)
      .then((routes) => setStatus({ kind: "done", routes }))
      .catch((err: unknown) => {
        if (controller.signal.aborted) return; // we cancelled it ourselves: not an error
        setStatus({ kind: "error", message: err instanceof Error ? err.message : String(err) });
      });
    // React runs this before the next effect and on unmount, so an old answer can't overwrite a new one.
    return () => controller.abort();
  }, [query, attempt]);

  return (
    <main>
      <h1>Where are you going?</h1>
      <label>
        Destination <input value={query} onChange={(e) => setQuery(e.target.value)} />
      </label>
      {status.kind === "loading" && <p role="status">Loading routes...</p>}
      {status.kind === "error" && (
        <div role="alert">
          Couldn't load routes: {status.message}{" "}
          <button type="button" onClick={() => setAttempt((a) => a + 1)}>
            Try again
          </button>
        </div>
      )}
      {status.kind === "done" && status.routes.length === 0 && <p>No routes match.</p>}
      {status.kind === "done" && status.routes.length > 0 && (
        <ul aria-label="Routes">
          {status.routes.map((r) => (
            <li key={`${r.operator}-${r.goesTo}`}>{label(r)}</li>
          ))}
        </ul>
      )}
    </main>
  );
}
