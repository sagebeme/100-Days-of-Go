// Day 66: your first React components. Each one is a function that takes props and returns JSX.
import type { Route } from "./routes";

/** FareBadge shows "Ksh100", or "Ask the conductor" when the fare isn't known. */
export function FareBadge({ amountKsh }: { amountKsh?: number }) {
  // Your code here
  return <span></span>;
}

/** RouteCard shows one route as an <article>: an <h2> like "58 to Buruburu", the operator, "Boards at ...", and a FareBadge. */
export function RouteCard({ route }: { route: Route }) {
  // Your code here
  return <article></article>;
}

/** App is the whole page: an <h1> "Matatu routes from the CBD" and a RouteCard for every route, or "No routes yet." */
export function App({ routes }: { routes: Route[] }) {
  // Your code here
  return <main></main>;
}
