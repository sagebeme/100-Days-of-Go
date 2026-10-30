// Day 74: a trip board that stays quick on a Ksh8,000 phone.
import { Suspense, lazy, useMemo, useState } from "react";
import { srcSet } from "./perf";
import { type Trip, searchTrips } from "./trips";

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
  // Your code here
  return <main></main>;
}
