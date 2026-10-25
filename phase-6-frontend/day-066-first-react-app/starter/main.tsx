// The entry point Vite loads from index.html. It finds <div id="root"> and hands it to React.
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import { ROUTES } from "./routes";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App routes={ROUTES} />
  </StrictMode>,
);
