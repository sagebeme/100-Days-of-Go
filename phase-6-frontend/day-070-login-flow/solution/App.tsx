// Day 70: a login page, a protected page, logout, and what to do when the session ends.
import { type FormEvent, type ReactNode, useCallback, useEffect, useState } from "react";
import { type AuthApi, SessionExpiredError, type User } from "./auth";

/** RequireAuth shows children only when there is a user, and fallback otherwise. */
export function RequireAuth({ user, fallback, children }: { user: User | null; fallback: ReactNode; children: ReactNode }) {
  return <>{user ? children : fallback}</>;
}

interface Booking {
  reference: string;
  route: string;
  seats: number;
}

const ENDED = "Your session has ended. Please log in again.";

function LoginForm({ auth, notice, onLogin }: { auth: AuthApi; notice: string; onLogin: (u: User) => void }) {
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [sending, setSending] = useState(false);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSending(true);
    setError("");
    try {
      onLogin(await auth.login(phone, password));
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
      setPassword(""); // never leave a wrong password sitting in the box
      setSending(false);
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <h2>Log in</h2>
      {notice && <p>{notice}</p>}
      <label htmlFor="phone">Phone</label>
      <input id="phone" type="tel" autoComplete="username" value={phone} onChange={(e) => setPhone(e.target.value)} />
      <label htmlFor="password">Password</label>
      <input
        id="password"
        type="password"
        autoComplete="current-password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
      />
      <button type="submit" disabled={sending}>
        Log in
      </button>
      {error && <p role="alert">{error}</p>}
    </form>
  );
}

function MyBookings({ auth, onEnded }: { auth: AuthApi; onEnded: () => void }) {
  const [bookings, setBookings] = useState<Booking[] | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;
    auth
      .apiFetch("/api/me/bookings")
      .then((res) => res.json() as Promise<Booking[]>)
      .then((list) => {
        if (!cancelled) setBookings(list);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        if (err instanceof SessionExpiredError) onEnded();
        else setError("Couldn't load your bookings.");
      });
    return () => {
      cancelled = true;
    };
  }, [auth, onEnded]);

  if (error) return <p role="alert">{error}</p>;
  if (bookings === null) return <p role="status">Loading bookings...</p>;
  return (
    <ul aria-label="Bookings">
      {bookings.map((b) => (
        <li key={b.reference}>
          {b.reference}: route {b.route}, {b.seats} seats
        </li>
      ))}
    </ul>
  );
}

/**
 * App runs auth.restore() on load, showing "Checking your session..." until it answers. Then:
 * - logged out: an <h2>Log in</h2> form with "Phone", "Password" (type="password") and a "Log in"
 *   button. A failed login shows the error in role="alert" and clears the password.
 * - logged in: "Habari, NAME", a "Log out" button, and the bookings from
 *   auth.apiFetch("/api/me/bookings") as "REF: route ROUTE, N seats".
 * When the session ends (auth.onExpired, or a SessionExpiredError from apiFetch) it returns to the
 * login form with "Your session has ended. Please log in again."
 */
export function App({ auth }: { auth: AuthApi }) {
  const [checking, setChecking] = useState(true);
  const [user, setUser] = useState<User | null>(null);
  const [notice, setNotice] = useState("");

  // useCallback with no dependencies keeps this the same function on every render, so the
  // effects that depend on it don't re-run.
  const ended = useCallback(() => {
    setUser(null);
    setNotice(ENDED);
  }, []);

  useEffect(() => {
    auth.onExpired = ended;
    let cancelled = false;
    auth.restore().then((u) => {
      if (cancelled) return;
      setUser(u);
      setChecking(false);
    });
    return () => {
      cancelled = true;
    };
  }, [auth, ended]);

  async function logOut() {
    await auth.logout();
    setUser(null);
    setNotice("");
  }

  if (checking) return <p role="status">Checking your session...</p>;
  return (
    <main>
      <h1>Twende</h1>
      <RequireAuth
        user={user}
        fallback={
          <LoginForm
            auth={auth}
            notice={notice}
            onLogin={(u) => {
              setNotice("");
              setUser(u);
            }}
          />
        }
      >
        <p>Habari, {user?.name}</p>
        <button type="button" onClick={logOut}>
          Log out
        </button>
        <MyBookings auth={auth} onEnded={ended} />
      </RequireAuth>
    </main>
  );
}
