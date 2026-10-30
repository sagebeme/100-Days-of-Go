// The entry point Vite loads from index.html. Start the Go API first (see the README).
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { createClient } from "./api";
import { App } from "./App";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App client={createClient()} />
  </StrictMode>,
);
