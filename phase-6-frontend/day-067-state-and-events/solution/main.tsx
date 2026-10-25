// The entry point Vite loads from index.html.
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { SeatCounter, StageQueue } from "./App";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <main>
      <h1>Stage manager</h1>
      <SeatCounter capacity={14} />
      <StageQueue />
    </main>
  </StrictMode>,
);
