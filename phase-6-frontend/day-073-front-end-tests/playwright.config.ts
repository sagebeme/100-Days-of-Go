// Playwright runs a real browser against the real dev server. verify doesn't run it (it needs a
// browser download), but you can. From phase-6-frontend:
//
//   npx playwright install chromium
//   npx playwright test --config day-073-front-end-tests/playwright.config.ts
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  outputDir: "./dist/test-results", // dist/ is gitignored
  use: { baseURL: "http://localhost:5173" },
  projects: [{ name: "phone", use: { ...devices["Pixel 7"] } }],
  webServer: {
    command: "npm run dev -- day-073-front-end-tests/solution",
    cwd: "..",
    url: "http://localhost:5173",
    reuseExistingServer: true,
  },
});
