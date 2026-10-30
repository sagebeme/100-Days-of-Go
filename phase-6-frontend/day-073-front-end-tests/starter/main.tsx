// The entry point Vite loads from index.html. With no API running, the page shows its error state;
// the Playwright spec in ../e2e fakes the API in the browser instead.
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { TripList } from "./TripList";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <TripList />
  </StrictMode>,
);
