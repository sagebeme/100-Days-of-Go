import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { BookingForm, SeatPicker } from "./App";

// Every query here is by role and accessible name: what a screen reader user hears.
// If a test can't find "the radio called Seat 3", neither can they.

describe("SeatPicker", () => {
  it("is a group of radio buttons named by its legend", () => {
    render(<SeatPicker seatCount={4} taken={[2]} value={null} onChange={() => {}} />);
    const group = screen.getByRole("group", { name: "Choose a seat" });
    const radios = within(group).getAllByRole("radio");
    expect(radios.map((r) => r.getAttribute("aria-label") ?? (r as HTMLInputElement).labels?.[0]?.textContent)).toEqual([
      "Seat 1",
      "Seat 2 (taken)",
      "Seat 3",
      "Seat 4",
    ]);
  });
  it("disables taken seats", () => {
    render(<SeatPicker seatCount={4} taken={[2]} value={null} onChange={() => {}} />);
    expect((screen.getByRole("radio", { name: "Seat 2 (taken)" }) as HTMLInputElement).disabled).toBe(true);
    expect((screen.getByRole("radio", { name: "Seat 3" }) as HTMLInputElement).disabled).toBe(false);
  });
  it("works from the keyboard: Space picks the focused seat", async () => {
    const picked: number[] = [];
    render(<SeatPicker seatCount={4} taken={[]} value={null} onChange={(n) => picked.push(n)} />);
    const user = userEvent.setup();
    await user.tab();
    expect(document.activeElement).toBe(screen.getByRole("radio", { name: "Seat 1" }));
    await user.keyboard(" ");
    expect(picked).toEqual([1]);
  });
});

describe("BookingForm", () => {
  function renderForm() {
    render(<BookingForm seatCount={6} taken={[2, 5]} />);
    return userEvent.setup();
  }

  it("labels every field", () => {
    renderForm();
    screen.getByRole("textbox", { name: "Name" });
    screen.getByRole("textbox", { name: "Phone" });
    screen.getByRole("button", { name: "Book" });
  });

  it("on a bad submit, moves focus to an error summary that links to each field", async () => {
    const user = renderForm();
    await user.click(screen.getByRole("button", { name: "Book" }));
    const summary = screen.getByRole("alert");
    expect(document.activeElement, "focus moves to the summary").toBe(summary);
    within(summary).getByRole("heading", { name: "There is a problem" });
    const links = within(summary).getAllByRole("link");
    expect(links.map((a) => a.textContent)).toEqual(["Enter your name", "Choose a seat"]);
    await user.click(links[0]);
    expect(document.activeElement, "the link moves focus to the field").toBe(screen.getByRole("textbox", { name: "Name" }));
  });

  it("ties each error to its field", async () => {
    const user = renderForm();
    await user.click(screen.getByRole("button", { name: "Book" }));
    const name = screen.getByRole("textbox", { name: "Name" });
    expect(name.getAttribute("aria-invalid")).toBe("true");
    expect(document.getElementById(name.getAttribute("aria-describedby") ?? "")?.textContent).toBe("Enter your name");
  });

  it("opens a dialog with focus inside, and Escape returns focus to Book", async () => {
    const user = renderForm();
    await user.type(screen.getByRole("textbox", { name: "Name" }), "Amina");
    await user.click(screen.getByRole("radio", { name: "Seat 3" }));
    await user.click(screen.getByRole("button", { name: "Book" }));
    const dialog = screen.getByRole("dialog", { name: "Confirm your booking" });
    expect(dialog.getAttribute("aria-modal")).toBe("true");
    within(dialog).getByText("Seat 3 for Amina");
    expect(document.activeElement).toBe(within(dialog).getByRole("button", { name: "Confirm" }));
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(document.activeElement).toBe(screen.getByRole("button", { name: "Book" }));
  });

  it("keeps Tab inside the dialog", async () => {
    const user = renderForm();
    await user.type(screen.getByRole("textbox", { name: "Name" }), "Amina");
    await user.click(screen.getByRole("radio", { name: "Seat 1" }));
    await user.click(screen.getByRole("button", { name: "Book" }));
    const dialog = screen.getByRole("dialog");
    const confirm = within(dialog).getByRole("button", { name: "Confirm" });
    const cancel = within(dialog).getByRole("button", { name: "Cancel" });
    await user.tab();
    expect(document.activeElement).toBe(cancel);
    await user.tab();
    expect(document.activeElement, "Tab from the last button wraps to the first").toBe(confirm);
    await user.tab({ shift: true });
    expect(document.activeElement, "Shift+Tab from the first wraps to the last").toBe(cancel);
  });

  it("confirms and announces the booking", async () => {
    const user = renderForm();
    await user.type(screen.getByRole("textbox", { name: "Name" }), "Amina");
    await user.click(screen.getByRole("radio", { name: "Seat 4" }));
    await user.click(screen.getByRole("button", { name: "Book" }));
    await user.click(screen.getByRole("button", { name: "Confirm" }));
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(screen.getByRole("status").textContent).toBe("Seat 4 is yours, Amina.");
  });
});
