// Made-up departures on real routes (checked on nishukishe.com, October 2026). Finished: use as is.

export interface Trip {
  id: number;
  route: string;
  goesTo: string;
  departs: string; // "07:45"
}

const ROUTES = [
  { route: "46K", goesTo: "Kawangware" },
  { route: "111", goesTo: "Ngong" },
  { route: "125", goesTo: "Ongata Rongai" },
  { route: "45", goesTo: "Githurai" },
  { route: "58", goesTo: "Buruburu" },
];

/** makeTrips makes n trips, one every two minutes from 05:00, cycling through the routes. */
export function makeTrips(n: number): Trip[] {
  return Array.from({ length: n }, (_, i) => {
    const minutes = 5 * 60 + i * 2;
    const hh = String(Math.floor(minutes / 60) % 24).padStart(2, "0");
    const mm = String(minutes % 60).padStart(2, "0");
    return { id: i + 1, ...ROUTES[i % ROUTES.length], departs: `${hh}:${mm}` };
  });
}

/** searchTrips keeps trips whose route or destination contains the query, ignoring case. */
export function searchTrips(trips: Trip[], query: string): Trip[] {
  const q = query.trim().toLowerCase();
  if (q === "") return trips;
  return trips.filter((t) => t.route.toLowerCase().includes(q) || t.goesTo.toLowerCase().includes(q));
}
