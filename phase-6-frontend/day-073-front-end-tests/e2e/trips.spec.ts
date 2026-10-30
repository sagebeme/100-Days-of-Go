// An end-to-end test: a real browser, the real page, and a fake API answered inside the browser.
// It's a .spec.ts file, so Vitest ignores it; Playwright runs it (see ../playwright.config.ts).
import { expect, test } from "@playwright/test";

const trips = [
  { id: "T1", route: "58", goesTo: "Buruburu", departsAt: "2099-01-01T07:42:00Z", seatsLeft: 3, fareKsh: 150 },
  { id: "T2", route: "111", goesTo: "Ngong", departsAt: "2099-01-01T07:35:00Z", seatsLeft: 0, fareKsh: 100 },
];

test("a passenger filters out full trips", async ({ page }) => {
  // Answer the API call in the browser, so the test needs no Go server.
  await page.route("**/api/trips", (route) => route.fulfill({ json: trips }));
  await page.goto("/");

  const list = page.getByRole("list", { name: "Trips" });
  await expect(list.getByRole("listitem")).toHaveCount(2);

  await page.getByRole("checkbox", { name: "Only trips with seats" }).check();
  await expect(list.getByRole("listitem")).toHaveCount(1);
  await expect(list).toContainText("58 to Buruburu");
});

test("the error state offers a retry", async ({ page }) => {
  // Fail until the error is on screen. (In development, StrictMode fetches twice on load, so
  // "fail the first call only" would be flaky.)
  let fail = true;
  await page.route("**/api/trips", (route) =>
    fail ? route.fulfill({ status: 503, json: { error: "down" } }) : route.fulfill({ json: trips }),
  );
  await page.goto("/");
  await expect(page.getByRole("alert")).toContainText("Couldn't load trips.");
  fail = false;
  await page.getByRole("button", { name: "Try again" }).click();
  await expect(page.getByRole("list", { name: "Trips" }).getByRole("listitem")).toHaveCount(2);
});
