import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { BookingForm } from "./App";

function reply(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}

async function fill(fields: { name?: string; phone?: string; route?: string; seats?: string }) {
  const user = userEvent.setup();
  if (fields.name !== undefined) await user.type(screen.getByLabelText("Name"), fields.name);
  if (fields.phone !== undefined) await user.type(screen.getByLabelText("Phone"), fields.phone);
  if (fields.route !== undefined) await user.selectOptions(screen.getByLabelText("Route"), fields.route);
  if (fields.seats !== undefined) {
    await user.clear(screen.getByLabelText("Seats"));
    await user.type(screen.getByLabelText("Seats"), fields.seats);
  }
  return user;
}

function submit() {
  return screen.getByRole("button", { name: /^Book/ });
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("BookingForm", () => {
  it("is a controlled form: what you type is what it holds", async () => {
    render(<BookingForm />);
    await fill({ name: "Amina", seats: "3" });
    expect((screen.getByLabelText("Name") as HTMLInputElement).value).toBe("Amina");
    expect((screen.getByLabelText("Seats") as HTMLInputElement).value).toBe("3");
  });

  it("shows the browser's checks next to each field and sends nothing", async () => {
    const fake = vi.fn();
    vi.stubGlobal("fetch", fake);
    render(<BookingForm />);
    const user = await fill({ phone: "123", seats: "0" });
    await user.click(submit());
    const name = screen.getByLabelText("Name");
    expect(name.getAttribute("aria-invalid")).toBe("true");
    const describedBy = name.getAttribute("aria-describedby");
    expect(describedBy, "the name input points at its error").toBeTruthy();
    expect(document.getElementById(describedBy!)?.textContent).toBe("Enter your name");
    screen.getByText("Enter a Kenyan mobile number like 0712 345 678");
    screen.getByText("Choose a route");
    screen.getByText("Book between 1 and 14 seats");
    expect(fake, "fetch is not called while the form has errors").not.toHaveBeenCalled();
  });

  it("sends the booking as JSON and shows the reference", async () => {
    const fake = vi.fn(async () => reply(201, { reference: "K7Q2M9XA", phone: "254712345678", route: "58", seats: 2 }));
    vi.stubGlobal("fetch", fake);
    render(<BookingForm />);
    const user = await fill({ name: " Amina ", phone: "0712 345 678", route: "58", seats: "2" });
    await user.click(submit());
    await screen.findByText("Booked! Reference K7Q2M9XA. We'll SMS 254712345678.");
    const [, init] = fake.mock.calls[0] as unknown as [string, RequestInit];
    expect(init.method).toBe("POST");
    expect(JSON.parse(init.body as string)).toEqual({ name: " Amina ", phone: "0712 345 678", route: "58", seats: 2 });
  });

  it("shows the server's 422 field errors under the right field", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => reply(422, { errors: { seats: "Only 3 seats left on 111" } })));
    render(<BookingForm />);
    const user = await fill({ name: "Kip", phone: "0712345678", route: "111", seats: "4" });
    await user.click(submit());
    const msg = await screen.findByText("Only 3 seats left on 111");
    const seats = screen.getByLabelText("Seats");
    expect(seats.getAttribute("aria-invalid")).toBe("true");
    expect(seats.getAttribute("aria-describedby")).toBe(msg.id);
  });

  it("disables the button while sending", async () => {
    let finish: (r: Response) => void = () => {};
    vi.stubGlobal("fetch", vi.fn(() => new Promise<Response>((resolve) => (finish = resolve))));
    render(<BookingForm />);
    const user = await fill({ name: "Kip", phone: "0712345678", route: "46K", seats: "1" });
    await user.click(submit());
    expect(submit().textContent).toBe("Booking...");
    expect((submit() as HTMLButtonElement).disabled).toBe(true);
    finish(reply(201, { reference: "AAAA1111", phone: "254712345678", route: "46K", seats: 1 }));
    await screen.findByText(/^Booked!/);
  });

  it("shows a general error for anything else", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => reply(500, { error: "boom" })));
    render(<BookingForm />);
    const user = await fill({ name: "Kip", phone: "0712345678", route: "46K", seats: "1" });
    await user.click(submit());
    const alert = await screen.findByRole("alert");
    expect(alert.textContent).toBe("Booking failed (HTTP 500). Please try again.");
  });
});
