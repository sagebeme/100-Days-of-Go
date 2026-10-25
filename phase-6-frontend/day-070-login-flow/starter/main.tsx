// The entry point Vite loads from index.html. Start the Go API first (see the README).
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import { Auth } from "./auth";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App auth={new Auth()} />
  </StrictMode>,
);
