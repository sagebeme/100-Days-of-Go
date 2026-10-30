// The entry point Vite loads from index.html.
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BookingForm } from "./App";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BookingForm taken={[2, 5, 6, 11]} seatCount={14} />
  </StrictMode>,
);
