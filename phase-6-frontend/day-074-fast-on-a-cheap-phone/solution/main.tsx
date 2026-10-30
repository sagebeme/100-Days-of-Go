// The entry point Vite loads from index.html.
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { TripBoard } from "./App";
import { makeTrips } from "./trips";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <TripBoard trips={makeTrips(600)} />
  </StrictMode>,
);
