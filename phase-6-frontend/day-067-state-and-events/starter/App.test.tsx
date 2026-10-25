import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { SeatCounter, StageQueue } from "./App";

function button(name: string): HTMLButtonElement {
  return screen.getByRole("button", { name }) as HTMLButtonElement;
}

describe("SeatCounter", () => {
  it("starts empty with Alight disabled", () => {
    render(<SeatCounter capacity={14} />);
    screen.getByText("0 of 14 seats taken");
    expect(button("Alight").disabled, "Alight disabled at 0").toBe(true);
    expect(button("Board").disabled, "Board disabled at 0").toBe(false);
  });
  it("counts up and down", async () => {
    const user = userEvent.setup();
    render(<SeatCounter capacity={14} />);
    await user.click(button("Board"));
    await user.click(button("Board"));
    await user.click(button("Board"));
    await user.click(button("Alight"));
    screen.getByText("2 of 14 seats taken");
  });
  it("stops at capacity and says it's full", async () => {
    const user = userEvent.setup();
    render(<SeatCounter capacity={2} />);
    expect(screen.queryByRole("status")).toBeNull();
    await user.click(button("Board"));
    await user.click(button("Board"));
    screen.getByText("2 of 2 seats taken");
    expect(button("Board").disabled, "Board disabled when full").toBe(true);
    expect(screen.getByRole("status").textContent).toBe("Imejaa! The mat is full.");
  });
});

describe("StageQueue", () => {
  async function addNames(...names: string[]) {
    const user = userEvent.setup();
    for (const n of names) {
      await user.type(screen.getByLabelText("Passenger name"), n);
      await user.click(button("Add to queue"));
    }
    return user;
  }

  it("adds names in order and clears the box", async () => {
    render(<StageQueue />);
    screen.getByText("In the queue: 0");
    await addNames("Amina", "Baraka");
    const items = within(screen.getByRole("list", { name: "Queue" })).getAllByRole("listitem");
    expect(items.map((li) => li.querySelector(".name")?.textContent)).toEqual(["Amina", "Baraka"]);
    screen.getByText("In the queue: 2");
    expect((screen.getByLabelText("Passenger name") as HTMLInputElement).value).toBe("");
  });
  it("ignores a blank name", async () => {
    render(<StageQueue />);
    await addNames("   ");
    screen.getByText("In the queue: 0");
  });
  it("adds on Enter, because the box is in a form", async () => {
    const user = userEvent.setup();
    render(<StageQueue />);
    await user.type(screen.getByLabelText("Passenger name"), "Chebet{Enter}");
    screen.getByText("In the queue: 1");
  });
  it("boards the right Kamau when two share a name", async () => {
    render(<StageQueue />);
    const user = await addNames("Kamau", "Kamau");
    await user.type(screen.getByLabelText("Stop for Kamau (2)"), "Ngong");
    const boardButtons = screen.getAllByRole("button", { name: "Board Kamau" });
    await user.click(boardButtons[0]);
    screen.getByText("In the queue: 1");
    const left = screen.getByLabelText("Stop for Kamau (2)") as HTMLInputElement;
    expect(left.value, "the second Kamau keeps what was typed for them").toBe("Ngong");
  });
  it("keeps each passenger's typed stop with that passenger (keys)", async () => {
    render(<StageQueue />);
    const user = await addNames("Amina", "Baraka");
    await user.type(screen.getByLabelText("Stop for Amina (1)"), "Kibera");
    await user.click(button("Board Amina"));
    const baraka = screen.getByLabelText("Stop for Baraka (2)") as HTMLInputElement;
    expect(baraka.value, "Baraka's stop after Amina boards").toBe("");
  });
});
