// Day 75: the page shows which environment it is, so nobody books a real seat on staging by mistake.
import type { Config } from "./config";

/**
 * EnvBadge shows <p role="note">STAGING</p> (or DEVELOPMENT) outside production, and nothing in production.
 */
export function EnvBadge({ mode }: { mode: Config["mode"] }) {
  // Your code here
  return null;
}

/** App shows the badge, a heading "Twende", and "API: URL" plus "Seat map: on/off". */
export function App({ config }: { config: Config }) {
  // Your code here
  return <main></main>;
}
