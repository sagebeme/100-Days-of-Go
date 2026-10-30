// Day 75: one place that reads, checks and explains the front end's configuration.
//
// Everything named VITE_* is pasted into the JavaScript at build time. Anyone can read it with
// View Source, so it must never hold a secret.

export type Mode = "development" | "staging" | "production";

export interface Config {
  mode: Mode;
  apiUrl: string; // no trailing slash
  wsUrl: string;
  seatMap: boolean; // feature flag: show the seat map
}

/** ConfigError lists every problem at once, so one deploy attempt shows them all. */
export class ConfigError extends Error {
  problems: string[];
  constructor(problems: string[]) {
    super(`Bad configuration:\n- ${problems.join("\n- ")}`);
    this.name = "ConfigError";
    this.problems = problems;
  }
}

/** Env is what import.meta.env holds: strings, or undefined when a variable isn't set. */
export type Env = Record<string, string | boolean | undefined>;

/**
 * parseConfig reads:
 * - VITE_APP_ENV: development (the default), staging or production.
 *   Otherwise: `VITE_APP_ENV must be development, staging or production (got "X")`
 * - VITE_API_URL: an http or https URL; a trailing slash is removed. If unset, development uses
 *   http://localhost:8080 and the others fail with `VITE_API_URL is required in MODE`.
 *   Not a URL: `VITE_API_URL must be an http or https URL (got "X")`.
 *   Outside development it must be https: `VITE_API_URL must use https in MODE`.
 * - VITE_WS_URL: a ws or wss URL (`VITE_WS_URL must be a ws or wss URL (got "X")`). If unset it
 *   comes from the API URL: http→ws, https→wss, plus "/ws/seats".
 * - VITE_SEAT_MAP: "true" or "false", default false. Otherwise `VITE_SEAT_MAP must be true or false (got "X")`.
 * Every problem found goes into one ConfigError, in the order above.
 */
export function parseConfig(env: Env): Config {
  // Your code here
  return { mode: "development", apiUrl: "", wsUrl: "", seatMap: false };
}

/**
 * findLeaks returns the names of VITE_ variables that look like secrets (their name contains
 * SECRET, PASSWORD, PRIVATE or TOKEN, any case), sorted. They would ship to every browser.
 */
export function findLeaks(env: Env): string[] {
  // Your code here
  return [];
}

/**
 * mergeRuntime applies /config.json on top of the build-time config. It may set apiUrl (string),
 * wsUrl (string) and seatMap (boolean); other keys are ignored. The result is checked with the same
 * rules as parseConfig (so apiUrl must be https outside development). If apiUrl is given without
 * wsUrl, wsUrl is worked out from the new apiUrl. A value of the wrong type, or JSON that isn't an
 * object, throws ConfigError.
 */
export function mergeRuntime(base: Config, runtime: unknown): Config {
  // Your code here
  return base;
}
