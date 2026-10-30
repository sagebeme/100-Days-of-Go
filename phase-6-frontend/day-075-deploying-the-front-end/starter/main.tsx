// The entry point. It reads the build-time config, then lets /config.json (if the host serves one)
// override it at runtime, and only then draws the page.
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import { type Config, mergeRuntime, parseConfig } from "./config";

async function start() {
  const root = createRoot(document.getElementById("root")!);
  let config: Config;
  try {
    config = parseConfig(import.meta.env);
    const res = await fetch("/config.json").catch(() => null);
    if (res?.ok && res.headers.get("Content-Type")?.includes("application/json")) {
      config = mergeRuntime(config, await res.json());
    }
  } catch (err) {
    // Fail loudly and early: a page pointing at the wrong API is worse than no page.
    root.render(<pre role="alert">{err instanceof Error ? err.message : String(err)}</pre>);
    return;
  }
  root.render(
    <StrictMode>
      <App config={config} />
    </StrictMode>,
  );
}

start();
