import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { App, FareBadge, RouteCard } from "./App";
import { ROUTES, type Route } from "./routes";

describe("FareBadge", () => {
  const cases: { name: string; amountKsh?: number; want: string }[] = [
    { name: "a fare", amountKsh: 100, want: "Ksh100" },
    { name: "a bigger fare", amountKsh: 150, want: "Ksh150" },
    { name: "no fare known", amountKsh: undefined, want: "Ask the conductor" },
  ];
  for (const c of cases) {
    it(c.name, () => {
      const { container } = render(<FareBadge amountKsh={c.amountKsh} />);
      expect(container.textContent, `<FareBadge amountKsh={${c.amountKsh}} />`).toBe(c.want);
    });
  }
});

describe("RouteCard", () => {
  it("shows the number, destination, operator, stage and fare", () => {
    render(<RouteCard route={ROUTES[4]} />);
    const card = screen.getByRole("article");
    expect(within(card).getByRole("heading", { level: 2 }).textContent).toBe("58 to Buruburu");
    within(card).getByText("Buruburu 58 Travellers Sacco");
    within(card).getByText("Boards at Ambassadeur");
    within(card).getByText("Ksh150");
  });
  it("leaves the number out when a route has none", () => {
    render(<RouteCard route={ROUTES[5]} />);
    expect(screen.getByRole("heading", { level: 2 }).textContent).toBe("To Kibera");
    screen.getByText("Ask the conductor");
  });
});

describe("App", () => {
  it("has a page heading and one card per route", () => {
    render(<App routes={ROUTES} />);
    screen.getByRole("heading", { level: 1, name: "Matatu routes from the CBD" });
    const cards = screen.getAllByRole("article");
    expect(cards.length, "number of route cards").toBe(6);
    expect(within(cards[0]).getByRole("heading").textContent).toBe("46K to Kawangware");
  });
  it("says so when there are no routes", () => {
    const none: Route[] = [];
    render(<App routes={none} />);
    expect(screen.queryAllByRole("article").length).toBe(0);
    screen.getByText("No routes yet.");
  });
});
