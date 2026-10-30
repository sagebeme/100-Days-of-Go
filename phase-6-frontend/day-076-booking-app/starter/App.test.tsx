import { act, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import {
  type Booking,
  type BookingInput,
  type BookingResult,
  type Client,
  LoginError,
  SessionExpiredError,
  type Trip,
  type User,
} from "./api";
import { App } from "./App";

const AMINA: User = { phone: "0712345678", name: "Amina" };
const TRIPS: Trip[] = [
  { id: "T1", route: "58", goesTo: "Buruburu", boardsAt: "Ambassadeur", departs: "07:45", fareKsh: 150, seatsLeft: 3 },
  { id: "T2", route: "111", goesTo: "Ngong", boardsAt: "Railways", departs: "07:50", fareKsh: 100, seatsLeft: 0 },
  { id: "T5", route: "45", goesTo: "Githurai", boardsAt: "Ronald Ngala Street", departs: "08:15", fareKsh: 100, seatsLeft: 1 },
];
const BOOKED: Booking = { reference: "K7Q2M9XA", tripId: "T1", seat: 3, phone: "254712345678", totalKsh: 150 };

/** fakeClient is a Client with no network. Override any method for one test. */
function fakeClient(over: Partial<Client> = {}): Client {
  return {
    onExpired: () => {},
    restore: vi.fn(async () => null),
    login: vi.fn(async (_phone: string, password: string) => {
      if (password !== "matatu-2026") throw new LoginError();
      return AMINA;
    }),
    logout: vi.fn(async () => {}),
    listTrips: vi.fn(async () => TRIPS),
    getSeats: vi.fn(async (tripId: string) => ({ tripId, seatCount: 6, taken: [1, 2, 4] })),
    book: vi.fn(async (_input: BookingInput): Promise<BookingResult> => ({ kind: "booked", booking: BOOKED })),
    ...over,
  };
}

async function openTrip58(user = userEvent.setup()) {
  await user.click(await screen.findByRole("button", { name: "Book the 58 at 07:45" }));
  await screen.findByRole("group", { name: "Choose a seat" });
  return user;
}

async function fillBooking(user: ReturnType<typeof userEvent.setup>, opts: { seat?: number; name?: string; phone?: string } = {}) {
  if (opts.seat !== undefined) await user.click(screen.getByRole("radio", { name: `Seat ${opts.seat}` }));
  const name = screen.getByRole("textbox", { name: "Name" });
  const phone = screen.getByRole("textbox", { name: "Phone" });
  await user.clear(name);
  await user.clear(phone);
  if (opts.name) await user.type(name, opts.name);
  if (opts.phone) await user.type(phone, opts.phone);
}

describe("the trips list", () => {
  it("loads and lists trips, with full ones not bookable", async () => {
    render(<App client={fakeClient()} />);
    screen.getByText("Loading trips...");
    const items = within(await screen.findByRole("list", { name: "Trips" })).getAllByRole("listitem");
    expect(items[0].textContent).toContain("58 to Buruburu, 07:45 from Ambassadeur, Ksh150, 3 seats left");
    expect(items[1].textContent).toContain("111 to Ngong, 07:50 from Railways, Ksh100, full");
    expect(items[2].textContent).toContain("1 seat left");
    expect((screen.getByRole("button", { name: "Book the 111 at 07:50" }) as HTMLButtonElement).disabled).toBe(true);
  });

  it("shows an error with Try again", async () => {
    const listTrips = vi.fn().mockRejectedValueOnce(new Error("database is down")).mockResolvedValueOnce(TRIPS);
    render(<App client={fakeClient({ listTrips })} />);
    const alert = await screen.findByRole("alert");
    expect(alert.textContent).toContain("Couldn't load trips: database is down");
    await userEvent.setup().click(within(alert).getByRole("button", { name: "Try again" }));
    await screen.findByRole("list", { name: "Trips" });
  });
});

describe("a trip's page", () => {
  it("shows the seat picker with taken seats disabled, and goes back", async () => {
    const client = fakeClient();
    render(<App client={client} />);
    const user = await openTrip58();
    screen.getByRole("heading", { name: "58 to Buruburu at 07:45" });
    expect(client.getSeats).toHaveBeenCalledWith("T1");
    expect((screen.getByRole("radio", { name: "Seat 2 (taken)" }) as HTMLInputElement).disabled).toBe(true);
    expect((screen.getByRole("radio", { name: "Seat 3" }) as HTMLInputElement).disabled).toBe(false);
    await user.click(screen.getByRole("button", { name: "Back to trips" }));
    await screen.findByRole("list", { name: "Trips" });
  });

  it("asks a logged-out passenger to log in, then brings them back", async () => {
    render(<App client={fakeClient()} />);
    const user = await openTrip58();
    expect(screen.queryByRole("button", { name: /^Book seat/ })).toBeNull();
    await user.click(screen.getByRole("button", { name: "Log in to book" }));
    screen.getByRole("heading", { name: "Log in" });
    await user.type(screen.getByLabelText("Phone"), "0712345678");
    await user.type(screen.getByLabelText("Password"), "matatu-2026");
    await user.click(screen.getByRole("button", { name: "Log in" }));
    await screen.findByRole("heading", { name: "58 to Buruburu at 07:45" });
    screen.getByRole("button", { name: "Book seat, Ksh150" });
    screen.getByText(/Habari, Amina/);
  });

  it("shows a wrong password and stays on the login page", async () => {
    render(<App client={fakeClient()} />);
    const user = userEvent.setup();
    await user.click(await screen.findByRole("button", { name: "Log in" }));
    await user.type(screen.getByLabelText("Phone"), "0712345678");
    await user.type(screen.getByLabelText("Password"), "1234");
    await user.click(screen.getByRole("button", { name: "Log in" }));
    expect((await screen.findByRole("alert")).textContent).toBe("Wrong phone number or password");
    expect((screen.getByLabelText("Password") as HTMLInputElement).value).toBe("");
  });
});

describe("booking", () => {
  const loggedIn = (over: Partial<Client> = {}) => fakeClient({ restore: vi.fn(async () => AMINA), ...over });

  it("checks the form in the browser first and sends nothing", async () => {
    const client = loggedIn();
    render(<App client={client} />);
    const user = await openTrip58();
    await fillBooking(user, { phone: "12" });
    await user.click(screen.getByRole("button", { name: "Book seat, Ksh150" }));
    const group = screen.getByRole("group", { name: "Choose a seat" });
    expect(document.getElementById(group.getAttribute("aria-describedby") ?? "")?.textContent, "the seat group's error").toBe(
      "Choose a seat",
    );
    screen.getByText("Enter your name");
    const phone = screen.getByRole("textbox", { name: "Phone" });
    expect(phone.getAttribute("aria-invalid")).toBe("true");
    expect(document.getElementById(phone.getAttribute("aria-describedby") ?? "")?.textContent).toBe(
      "Enter a Kenyan mobile number like 0712 345 678",
    );
    expect(client.book).not.toHaveBeenCalled();
  });

  it("books and shows the confirmation", async () => {
    const client = loggedIn();
    render(<App client={client} />);
    const user = await openTrip58();
    await fillBooking(user, { seat: 3, name: " Amina ", phone: "0712 345 678" });
    await user.click(screen.getByRole("button", { name: "Book seat, Ksh150" }));
    await screen.findByRole("heading", { name: "Booked!" });
    screen.getByText("Reference K7Q2M9XA. Seat 3 on the 58 to Buruburu at 07:45. Ksh150. We'll SMS 254712345678.");
    expect(client.book).toHaveBeenCalledWith({ tripId: "T1", seat: 3, name: "Amina", phone: "0712 345 678" });
    await user.click(screen.getByRole("button", { name: "Book another trip" }));
    await screen.findByRole("list", { name: "Trips" });
  });

  it("shows the server's field errors", async () => {
    const book = vi.fn(async (): Promise<BookingResult> => ({ kind: "invalid", errors: { name: "Name is too long (50 characters at most)" } }));
    render(<App client={loggedIn({ book })} />);
    const user = await openTrip58();
    await fillBooking(user, { seat: 3, name: "Amina", phone: "0712345678" });
    await user.click(screen.getByRole("button", { name: "Book seat, Ksh150" }));
    await screen.findByText("Name is too long (50 characters at most)");
    expect(screen.getByRole("textbox", { name: "Name" }).getAttribute("aria-invalid")).toBe("true");
  });

  it("handles a seat someone else just took", async () => {
    const book = vi.fn(async (): Promise<BookingResult> => ({ kind: "conflict", message: "Seat 3 was just taken" }));
    const getSeats = vi
      .fn()
      .mockResolvedValueOnce({ tripId: "T1", seatCount: 6, taken: [1, 2, 4] })
      .mockResolvedValue({ tripId: "T1", seatCount: 6, taken: [1, 2, 3, 4] });
    render(<App client={loggedIn({ book, getSeats })} />);
    const user = await openTrip58();
    await fillBooking(user, { seat: 3, name: "Amina", phone: "0712345678" });
    await user.click(screen.getByRole("button", { name: "Book seat, Ksh150" }));
    expect((await screen.findByRole("alert")).textContent).toBe("Seat 3 was just taken. Pick another seat.");
    await screen.findByRole("radio", { name: "Seat 3 (taken)" });
    expect(getSeats, "the seat map is fetched again").toHaveBeenCalledTimes(2);
  });

  it("shows any other failure", async () => {
    const book = vi.fn(async (): Promise<BookingResult> => {
      throw new Error("HTTP 500");
    });
    render(<App client={loggedIn({ book })} />);
    const user = await openTrip58();
    await fillBooking(user, { seat: 3, name: "Amina", phone: "0712345678" });
    await user.click(screen.getByRole("button", { name: "Book seat, Ksh150" }));
    expect((await screen.findByRole("alert")).textContent).toBe("Booking failed: HTTP 500");
  });

  it("sends the passenger to log in when the session ends mid-booking", async () => {
    let client: Client;
    const book = vi.fn(async (): Promise<BookingResult> => {
      client.onExpired(); // what the real client does before throwing
      throw new SessionExpiredError();
    });
    client = loggedIn({ book });
    render(<App client={client} />);
    const user = await openTrip58();
    await fillBooking(user, { seat: 3, name: "Amina", phone: "0712345678" });
    await user.click(screen.getByRole("button", { name: "Book seat, Ksh150" }));
    await screen.findByRole("heading", { name: "Log in" });
    screen.getByText("Your session has ended. Please log in again.");
    expect(screen.queryByText(/Habari/)).toBeNull();
  });

  it("logs out", async () => {
    const client = loggedIn();
    render(<App client={client} />);
    const user = userEvent.setup();
    await user.click(await screen.findByRole("button", { name: "Log out" }));
    await screen.findByRole("button", { name: "Log in" });
    expect(client.logout).toHaveBeenCalledTimes(1);
  });

  it("moves focus to each new page's heading", async () => {
    render(<App client={loggedIn()} />);
    await openTrip58();
    expect(document.activeElement).toBe(screen.getByRole("heading", { name: "58 to Buruburu at 07:45" }));
  });
});

describe("the session", () => {
  it("restores a logged-in passenger on load", async () => {
    render(<App client={fakeClient({ restore: vi.fn(async () => AMINA) })} />);
    await screen.findByText(/Habari, Amina/);
  });

  it("logs out everywhere when onExpired is called", async () => {
    const client = fakeClient({ restore: vi.fn(async () => AMINA) });
    render(<App client={client} />);
    await screen.findByText(/Habari, Amina/);
    act(() => client.onExpired());
    await screen.findByText("Your session has ended. Please log in again.");
  });
});
