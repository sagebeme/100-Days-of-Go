// The entry point Vite loads from index.html. Start the Go API first (see the README).
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { LiveSeats } from "./App";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <LiveSeats />
  </StrictMode>,
);
