// @vitest-environment node
import { describe, expect, it } from "vitest";
import { type Config, ConfigError, type Env, findLeaks, mergeRuntime, parseConfig } from "./config";

function problems(fn: () => unknown): string[] {
  try {
    fn();
  } catch (err) {
    expect(err).toBeInstanceOf(ConfigError);
    return (err as ConfigError).problems;
  }
  throw new Error("expected a ConfigError, got none");
}

describe("parseConfig", () => {
  const good: { name: string; env: Env; want: Config }[] = [
    {
      name: "nothing set: development defaults",
      env: {},
      want: { mode: "development", apiUrl: "http://localhost:8080", wsUrl: "ws://localhost:8080/ws/seats", seatMap: false },
    },
    {
      name: "production with https, trailing slash removed",
      env: { VITE_APP_ENV: "production", VITE_API_URL: "https://api.twende.example/", VITE_SEAT_MAP: "true" },
      want: { mode: "production", apiUrl: "https://api.twende.example", wsUrl: "wss://api.twende.example/ws/seats", seatMap: true },
    },
    {
      name: "staging with its own WebSocket host",
      env: { VITE_APP_ENV: "staging", VITE_API_URL: "https://staging.twende.example", VITE_WS_URL: "wss://live.staging.twende.example/seats" },
      want: { mode: "staging", apiUrl: "https://staging.twende.example", wsUrl: "wss://live.staging.twende.example/seats", seatMap: false },
    },
    {
      name: "Vite's own keys are ignored",
      env: { MODE: "production", DEV: false, PROD: true, BASE_URL: "/", VITE_SEAT_MAP: "false" },
      want: { mode: "development", apiUrl: "http://localhost:8080", wsUrl: "ws://localhost:8080/ws/seats", seatMap: false },
    },
  ];
  for (const c of good) {
    it(c.name, () => {
      expect(parseConfig(c.env)).toEqual(c.want);
    });
  }

  const bad: { name: string; env: Env; want: string[] }[] = [
    {
      name: "production without an API URL",
      env: { VITE_APP_ENV: "production" },
      want: ["VITE_API_URL is required in production"],
    },
    {
      name: "plain http in staging",
      env: { VITE_APP_ENV: "staging", VITE_API_URL: "http://staging.twende.example" },
      want: ["VITE_API_URL must use https in staging"],
    },
    {
      name: "every problem at once",
      env: { VITE_APP_ENV: "prod", VITE_API_URL: "api.twende.example", VITE_WS_URL: "https://x", VITE_SEAT_MAP: "yes" },
      want: [
        'VITE_APP_ENV must be development, staging or production (got "prod")',
        'VITE_API_URL must be an http or https URL (got "api.twende.example")',
        'VITE_WS_URL must be a ws or wss URL (got "https://x")',
        'VITE_SEAT_MAP must be true or false (got "yes")',
      ],
    },
  ];
  for (const c of bad) {
    it(`rejects ${c.name}`, () => {
      expect(problems(() => parseConfig(c.env))).toEqual(c.want);
    });
  }

  it("puts every problem in the message", () => {
    expect(() => parseConfig({ VITE_APP_ENV: "staging" })).toThrow("Bad configuration:\n- VITE_API_URL is required in staging");
  });
});

describe("findLeaks", () => {
  it("flags VITE_ names that look secret, and only VITE_ ones", () => {
    const env: Env = {
      VITE_API_URL: "https://api.twende.example",
      VITE_MPESA_CONSUMER_SECRET: "oops",
      VITE_admin_password: "oops",
      VITE_GITHUB_TOKEN: "oops",
      JWT_SECRET: "fine: not VITE_, so Vite never ships it",
      VITE_MAPS_PUBLIC_KEY: "public by design",
    };
    expect(findLeaks(env)).toEqual(["VITE_GITHUB_TOKEN", "VITE_MPESA_CONSUMER_SECRET", "VITE_admin_password"]);
  });
});

describe("mergeRuntime", () => {
  const base = parseConfig({ VITE_APP_ENV: "production", VITE_API_URL: "https://api.twende.example" });

  it("lets one build point at another API at runtime", () => {
    expect(mergeRuntime(base, { apiUrl: "https://staging-api.twende.example", seatMap: true, colour: "green" })).toEqual({
      mode: "production",
      apiUrl: "https://staging-api.twende.example",
      wsUrl: "wss://staging-api.twende.example/ws/seats",
      seatMap: true,
    });
  });
  it("keeps the build-time values when /config.json is empty", () => {
    expect(mergeRuntime(base, {})).toEqual(base);
  });
  it("checks runtime values with the same rules", () => {
    expect(problems(() => mergeRuntime(base, { apiUrl: "http://api.twende.example" }))).toEqual([
      "VITE_API_URL must use https in production",
    ]);
  });
  it("rejects the wrong types", () => {
    expect(problems(() => mergeRuntime(base, { apiUrl: 42, seatMap: "yes" }))).toEqual([
      "apiUrl in /config.json must be a string",
      "seatMap in /config.json must be true or false",
    ]);
    expect(problems(() => mergeRuntime(base, ["not", "an", "object"]))).toEqual(["/config.json must be a JSON object"]);
  });
});
