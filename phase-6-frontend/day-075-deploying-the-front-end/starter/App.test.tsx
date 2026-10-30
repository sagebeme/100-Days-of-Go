import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { App, EnvBadge } from "./App";

describe("EnvBadge", () => {
  it("labels staging and development", () => {
    const { rerender } = render(<EnvBadge mode="staging" />);
    expect(screen.getByRole("note").textContent).toBe("STAGING");
    rerender(<EnvBadge mode="development" />);
    expect(screen.getByRole("note").textContent).toBe("DEVELOPMENT");
  });
  it("shows nothing in production", () => {
    const { container } = render(<EnvBadge mode="production" />);
    expect(container.textContent).toBe("");
  });
});

describe("App", () => {
  it("shows the config it was given", () => {
    render(<App config={{ mode: "staging", apiUrl: "https://staging.twende.example", wsUrl: "wss://x", seatMap: true }} />);
    screen.getByRole("heading", { name: "Twende" });
    screen.getByText("API: https://staging.twende.example");
    screen.getByText("Seat map: on");
    screen.getByRole("note");
  });
});
