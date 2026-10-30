import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TripBoard } from "./App";
import { makeTrips, searchTrips } from "./trips";

describe("TripBoard", () => {
  it("draws at most 50 rows, however many trips there are", () => {
    render(<TripBoard trips={makeTrips(600)} />);
    const rows = within(screen.getByRole("list", { name: "Trips" })).getAllByRole("listitem");
    expect(rows).toHaveLength(50);
    expect(rows[0].textContent).toBe("05:00 46K to Kawangware");
    screen.getByText("Showing 50 of 600 trips");
  });

  it("searches", async () => {
    render(<TripBoard trips={makeTrips(20)} />);
    await userEvent.setup().type(screen.getByRole("searchbox", { name: "Search trips" }), "ngong");
    const rows = within(screen.getByRole("list", { name: "Trips" })).getAllByRole("listitem");
    expect(rows.map((r) => r.textContent)).toEqual(["05:02 111 to Ngong", "05:12 111 to Ngong", "05:22 111 to Ngong", "05:32 111 to Ngong"]);
    screen.getByText("Showing 4 of 4 trips");
  });

  it("only searches again when the trips or the query change (useMemo)", async () => {
    const search = vi.fn(searchTrips);
    render(<TripBoard trips={makeTrips(20)} search={search} />);
    const user = userEvent.setup();
    await user.type(screen.getByRole("searchbox", { name: "Search trips" }), "58");
    const afterTyping = search.mock.calls.length;
    expect(afterTyping, "once on load and once per keystroke").toBe(3);
    await user.click(screen.getByRole("checkbox", { name: "Show seat map" }));
    await screen.findByRole("heading", { name: "Seat map" });
    await user.click(screen.getByRole("checkbox", { name: "Show seat map" }));
    expect(search.mock.calls.length, "toggling the seat map doesn't search again").toBe(afterTyping);
  });

  it("loads the seat map only when asked (React.lazy)", async () => {
    render(<TripBoard trips={makeTrips(5)} />);
    expect(screen.queryByRole("heading", { name: "Seat map" })).toBeNull();
    await userEvent.setup().click(screen.getByRole("checkbox", { name: "Show seat map" }));
    await screen.findByRole("heading", { name: "Seat map" });
  });

  it("serves a right-sized, lazy-loading picture", () => {
    render(<TripBoard trips={makeTrips(5)} />);
    const img = screen.getByRole("img", { name: "A route 58 matatu at Ambassadeur" }) as HTMLImageElement;
    expect(img.getAttribute("srcset")).toBe("/img/matatu-320.webp 320w, /img/matatu-640.webp 640w, /img/matatu-960.webp 960w");
    expect(img.getAttribute("sizes")).toBe("(max-width: 600px) 100vw, 600px");
    expect(img.getAttribute("loading")).toBe("lazy");
    expect(img.getAttribute("width"), "width and height stop the page jumping").toBe("640");
    expect(img.getAttribute("height")).toBe("360");
  });
});
