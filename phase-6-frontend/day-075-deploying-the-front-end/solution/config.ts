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

const MODES: Mode[] = ["development", "staging", "production"];

function isUrl(s: string, protocols: string[]): boolean {
  try {
    return protocols.includes(new URL(s).protocol);
  } catch {
    return false;
  }
}

function str(v: string | boolean | undefined): string | undefined {
  return v === undefined || v === "" ? undefined : String(v);
}

/** check applies the rules to finished values and returns the problems. parseConfig and mergeRuntime share it. */
function check(mode: Mode, apiUrl: string, wsUrl: string): string[] {
  const problems: string[] = [];
  if (!isUrl(apiUrl, ["http:", "https:"])) {
    problems.push(`VITE_API_URL must be an http or https URL (got "${apiUrl}")`);
  } else if (mode !== "development" && !apiUrl.startsWith("https:")) {
    problems.push(`VITE_API_URL must use https in ${mode}`);
  }
  if (!isUrl(wsUrl, ["ws:", "wss:"])) {
    problems.push(`VITE_WS_URL must be a ws or wss URL (got "${wsUrl}")`);
  }
  return problems;
}

function wsFrom(apiUrl: string): string {
  return apiUrl.replace(/^http/, "ws") + "/ws/seats"; // http→ws, https→wss
}

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
  const problems: string[] = [];

  const rawMode = str(env.VITE_APP_ENV) ?? "development";
  let mode: Mode = "development";
  if ((MODES as string[]).includes(rawMode)) {
    mode = rawMode as Mode;
  } else {
    problems.push(`VITE_APP_ENV must be development, staging or production (got "${rawMode}")`);
  }

  let apiUrl = str(env.VITE_API_URL)?.replace(/\/+$/, "");
  if (apiUrl === undefined) {
    if (mode === "development") {
      apiUrl = "http://localhost:8080";
    } else {
      problems.push(`VITE_API_URL is required in ${mode}`);
    }
  }

  let wsUrl = str(env.VITE_WS_URL);
  if (apiUrl !== undefined) {
    wsUrl ??= wsFrom(apiUrl);
    problems.push(...check(mode, apiUrl, wsUrl));
  } else if (wsUrl !== undefined && !isUrl(wsUrl, ["ws:", "wss:"])) {
    problems.push(`VITE_WS_URL must be a ws or wss URL (got "${wsUrl}")`);
  }

  const rawFlag = str(env.VITE_SEAT_MAP) ?? "false";
  if (rawFlag !== "true" && rawFlag !== "false") {
    problems.push(`VITE_SEAT_MAP must be true or false (got "${rawFlag}")`);
  }

  if (problems.length > 0) throw new ConfigError(problems);
  return { mode, apiUrl: apiUrl!, wsUrl: wsUrl!, seatMap: rawFlag === "true" };
}

/**
 * findLeaks returns the names of VITE_ variables that look like secrets (their name contains
 * SECRET, PASSWORD, PRIVATE or TOKEN, any case), sorted. They would ship to every browser.
 */
export function findLeaks(env: Env): string[] {
  return Object.keys(env)
    .filter((k) => k.startsWith("VITE_") && /SECRET|PASSWORD|PRIVATE|TOKEN/i.test(k))
    .sort();
}

/**
 * mergeRuntime applies /config.json on top of the build-time config. It may set apiUrl (string),
 * wsUrl (string) and seatMap (boolean); other keys are ignored. The result is checked with the same
 * rules as parseConfig (so apiUrl must be https outside development). If apiUrl is given without
 * wsUrl, wsUrl is worked out from the new apiUrl. A value of the wrong type, or JSON that isn't an
 * object, throws ConfigError.
 */
export function mergeRuntime(base: Config, runtime: unknown): Config {
  if (typeof runtime !== "object" || runtime === null || Array.isArray(runtime)) {
    throw new ConfigError(["/config.json must be a JSON object"]);
  }
  const r = runtime as Record<string, unknown>;
  const problems: string[] = [];
  const next = { ...base };
  for (const key of ["apiUrl", "wsUrl"] as const) {
    if (r[key] === undefined) continue;
    if (typeof r[key] === "string") next[key] = r[key].replace(/\/+$/, "");
    else problems.push(`${key} in /config.json must be a string`);
  }
  if (r.seatMap !== undefined) {
    if (typeof r.seatMap === "boolean") next.seatMap = r.seatMap;
    else problems.push("seatMap in /config.json must be true or false");
  }
  if (r.apiUrl !== undefined && r.wsUrl === undefined && typeof r.apiUrl === "string") {
    next.wsUrl = wsFrom(next.apiUrl); // a new API host means a new WebSocket host
  }
  problems.push(...check(next.mode, next.apiUrl, next.wsUrl));
  if (problems.length > 0) throw new ConfigError(problems);
  return next;
}
