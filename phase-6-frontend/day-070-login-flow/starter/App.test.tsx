import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { App, RequireAuth } from "./App";
import { type AuthApi, LoginError, SessionExpiredError, type User } from "./auth";

const AMINA: User = { phone: "0712345678", name: "Amina" };

/** fakeAuth is an AuthApi with no network. Pass what restore gives and how bookings answer. */
function fakeAuth(opts: { restored?: User | null; bookings?: () => Promise<Response> } = {}) {
  const auth: AuthApi = {
    onExpired: () => {},
    restore: vi.fn(async () => opts.restored ?? null),
    login: vi.fn(async (_phone: string, password: string) => {
      if (password !== "matatu-2026") throw new LoginError();
      return AMINA;
    }),
    logout: vi.fn(async () => {}),
    apiFetch: vi.fn(
      opts.bookings ??
        (async () => new Response(JSON.stringify([{ reference: "K7Q2M9XA", route: "58", seats: 2 }]), { status: 200 })),
    ),
  };
  return auth;
}

async function logIn(password = "matatu-2026") {
  const user = userEvent.setup();
  await user.type(await screen.findByLabelText("Phone"), "0712345678");
  await user.type(screen.getByLabelText("Password"), password);
  await user.click(screen.getByRole("button", { name: "Log in" }));
  return user;
}

describe("RequireAuth", () => {
  it("shows its children only to a logged-in user", () => {
    const { rerender } = render(
      <RequireAuth user={null} fallback={<p>Please log in</p>}>
        <p>Secret bookings</p>
      </RequireAuth>,
    );
    screen.getByText("Please log in");
    expect(screen.queryByText("Secret bookings")).toBeNull();
    rerender(
      <RequireAuth user={AMINA} fallback={<p>Please log in</p>}>
        <p>Secret bookings</p>
      </RequireAuth>,
    );
    screen.getByText("Secret bookings");
  });
});

describe("App", () => {
  it("checks the session first, then shows the login form", async () => {
    render(<App auth={fakeAuth()} />);
    screen.getByText("Checking your session...");
    await screen.findByRole("heading", { name: "Log in" });
  });

  it("goes straight to the bookings when the cookie still works", async () => {
    render(<App auth={fakeAuth({ restored: AMINA })} />);
    await screen.findByText("Habari, Amina");
    await screen.findByText("K7Q2M9XA: route 58, 2 seats");
  });

  it("logs in and shows the bookings", async () => {
    const auth = fakeAuth();
    render(<App auth={auth} />);
    await logIn();
    await screen.findByText("Habari, Amina");
    await screen.findByText("K7Q2M9XA: route 58, 2 seats");
    expect(auth.apiFetch).toHaveBeenCalledWith("/api/me/bookings");
  });

  it("shows a wrong password and clears the password box", async () => {
    render(<App auth={fakeAuth()} />);
    await logIn("1234");
    expect((await screen.findByRole("alert")).textContent).toBe("Wrong phone number or password");
    expect((screen.getByLabelText("Password") as HTMLInputElement).value).toBe("");
    expect((screen.getByLabelText("Password") as HTMLInputElement).type).toBe("password");
  });

  it("logs out back to the login form", async () => {
    const auth = fakeAuth({ restored: AMINA });
    render(<App auth={auth} />);
    await userEvent.setup().click(await screen.findByRole("button", { name: "Log out" }));
    await screen.findByRole("heading", { name: "Log in" });
    expect(auth.logout).toHaveBeenCalledTimes(1);
  });

  it("sends the passenger back to log in when the session ends", async () => {
    const auth = fakeAuth({ restored: AMINA, bookings: async () => Promise.reject(new SessionExpiredError()) });
    render(<App auth={auth} />);
    await screen.findByRole("heading", { name: "Log in" });
    screen.getByText("Your session has ended. Please log in again.");
  });

  it("sets auth.onExpired so a 401 anywhere logs the passenger out", async () => {
    const auth = fakeAuth({ restored: AMINA });
    render(<App auth={auth} />);
    await screen.findByText("Habari, Amina");
    act(() => auth.onExpired());
    await screen.findByText("Your session has ended. Please log in again.");
  });
});
