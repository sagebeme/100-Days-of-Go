// One config for every day. `npm run dev -- <day>/starter` serves that folder's index.html,
// and Vitest reads the `test` block below.
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  server: { port: 5173, strictPort: true },
  test: {
    environment: "jsdom",
    include: ["day-*/**/*.test.{ts,tsx}", ".verify/**/*.test.{ts,tsx}"],
    setupFiles: ["./vitest.setup.ts"],
  },
});
