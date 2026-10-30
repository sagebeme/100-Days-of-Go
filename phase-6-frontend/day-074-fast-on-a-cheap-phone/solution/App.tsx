// Day 74: a trip board that stays quick on a Ksh8,000 phone.
import { Suspense, lazy, useMemo, useState } from "react";
import { srcSet } from "./perf";
import { type Trip, searchTrips } from "./trips";

// Vite puts SeatMap in its own file, downloaded the first time it's shown.
const SeatMap = lazy(() => import("./SeatMap"));

const MAX_ROWS = 50;

/**
 * TripBoard shows:
 * - an <img> with alt "A route 58 matatu at Ambassadeur", srcSet from srcSet("matatu", [320, 640, 960]),
 *   sizes "(max-width: 600px) 100vw, 600px", loading="lazy", width 640 and height 360
 * - an <input type="search"> labelled "Search trips"; results come from search(trips, query),
 *   worked out with useMemo so they're only recomputed when trips or query change
 * - "Showing N of M trips" and <ul aria-label="Trips"> with at most 50 rows like "05:00 46K to Kawangware"
 * - a checkbox "Show seat map" that shows SeatMap, loaded with React.lazy inside <Suspense>
 */
export function TripBoard({
  trips,
  search = searchTrips,
}: {
  trips: Trip[];
  search?: (trips: Trip[], query: string) => Trip[];
}) {
  const [query, setQuery] = useState("");
  const [showMap, setShowMap] = useState(false);

  // Without useMemo, every render (even ticking the checkbox) would filter all the trips again.
  const found = useMemo(() => search(trips, query), [search, trips, query]);
  const shown = found.slice(0, MAX_ROWS); // drawing 600 rows costs more than searching them

  return (
    <main>
      <h1>Trip board</h1>
      <img
        src="/img/matatu-640.webp"
        srcSet={srcSet("matatu", [320, 640, 960])}
        sizes="(max-width: 600px) 100vw, 600px"
        alt="A route 58 matatu at Ambassadeur"
        width={640}
        height={360}
        loading="lazy"
        decoding="async"
      />
      <label>
        Search trips <input type="search" value={query} onChange={(e) => setQuery(e.target.value)} />
      </label>
      <label>
        <input type="checkbox" checked={showMap} onChange={(e) => setShowMap(e.target.checked)} /> Show seat map
      </label>
      {showMap && (
        <Suspense fallback={<p role="status">Loading seat map...</p>}>
          <SeatMap />
        </Suspense>
      )}
      <p>
        Showing {shown.length} of {found.length} trips
      </p>
      <ul aria-label="Trips">
        {shown.map((t) => (
          <li key={t.id}>
            {t.departs} {t.route} to {t.goesTo}
          </li>
        ))}
      </ul>
    </main>
  );
}
