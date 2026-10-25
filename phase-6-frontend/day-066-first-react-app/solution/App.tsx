// Day 66: your first React components. Each one is a function that takes props and returns JSX.
import type { Route } from "./routes";

/** FareBadge shows "Ksh100", or "Ask the conductor" when the fare isn't known. */
export function FareBadge({ amountKsh }: { amountKsh?: number }) {
  return <span className="fare">{amountKsh === undefined ? "Ask the conductor" : `Ksh${amountKsh}`}</span>;
}

/** RouteCard shows one route as an <article>: an <h2> like "58 to Buruburu", the operator, "Boards at ...", and a FareBadge. */
export function RouteCard({ route }: { route: Route }) {
  const title = route.number ? `${route.number} to ${route.goesTo}` : `To ${route.goesTo}`;
  return (
    <article>
      <h2>{title}</h2>
      <p>{route.operator}</p>
      <p>Boards at {route.boardsAt}</p>
      <FareBadge amountKsh={route.fareKsh} />
    </article>
  );
}

/** App is the whole page: an <h1> "Matatu routes from the CBD" and a RouteCard for every route, or "No routes yet." */
export function App({ routes }: { routes: Route[] }) {
  return (
    <main>
      <h1>Matatu routes from the CBD</h1>
      {routes.length === 0 ? (
        <p>No routes yet.</p>
      ) : (
        <ul>
          {routes.map((r) => (
            <li key={`${r.operator}-${r.goesTo}`}>
              <RouteCard route={r} />
            </li>
          ))}
        </ul>
      )}
      <p>Routes change, so confirm with the conductor.</p>
    </main>
  );
}
