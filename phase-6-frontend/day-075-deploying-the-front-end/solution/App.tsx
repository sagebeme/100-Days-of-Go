// Day 75: the page shows which environment it is, so nobody books a real seat on staging by mistake.
import type { Config } from "./config";

/**
 * EnvBadge shows <p role="note">STAGING</p> (or DEVELOPMENT) outside production, and nothing in production.
 */
export function EnvBadge({ mode }: { mode: Config["mode"] }) {
  if (mode === "production") return null;
  return (
    <p role="note" className={`env-badge env-${mode}`}>
      {mode.toUpperCase()}
    </p>
  );
}

/** App shows the badge, a heading "Twende", and "API: URL" plus "Seat map: on/off". */
export function App({ config }: { config: Config }) {
  return (
    <main>
      <EnvBadge mode={config.mode} />
      <h1>Twende</h1>
      <p>API: {config.apiUrl}</p>
      <p>Seat map: {config.seatMap ? "on" : "off"}</p>
    </main>
  );
}
