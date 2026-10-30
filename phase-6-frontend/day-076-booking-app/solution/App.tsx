// Day 76 capstone: the whole booking app.
import { type FormEvent, useCallback, useEffect, useRef, useState } from "react";
import {
  type Booking,
  type Client,
  type FieldErrors,
  type SeatInfo,
  SessionExpiredError,
  type Trip,
  type User,
} from "./api";
import { validateBooking } from "./validate";

type View =
  | { name: "trips" }
  | { name: "trip"; trip: Trip }
  | { name: "login"; back: View; notice: string }
  | { name: "done"; trip: Trip; booking: Booking };

const ENDED = "Your session has ended. Please log in again.";

function title(t: Trip): string {
  return `${t.route} to ${t.goesTo} at ${t.departs}`;
}

function message(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

/** A heading that takes focus when it appears, so screen reader users hear the new page (Day 72). */
function PageHeading({ children }: { children: string }) {
  const ref = useRef<HTMLHeadingElement>(null);
  useEffect(() => ref.current?.focus(), []);
  return (
    <h2 tabIndex={-1} ref={ref}>
      {children}
    </h2>
  );
}

function TripsPage({ client, onPick }: { client: Client; onPick: (t: Trip) => void }) {
  const [trips, setTrips] = useState<Trip[] | null>(null);
  const [error, setError] = useState("");
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setTrips(null);
    setError("");
    client.listTrips().then(
      (list) => !cancelled && setTrips(list),
      (err: unknown) => !cancelled && setError(message(err)),
    );
    return () => {
      cancelled = true;
    };
  }, [client, attempt]);

  return (
    <section>
      <PageHeading>Next trips</PageHeading>
      {error && (
        <div role="alert">
          Couldn't load trips: {error}{" "}
          <button type="button" onClick={() => setAttempt((a) => a + 1)}>
            Try again
          </button>
        </div>
      )}
      {!error && trips === null && <p role="status">Loading trips...</p>}
      {trips && (
        <ul aria-label="Trips">
          {trips.map((t) => (
            <li key={t.id}>
              {t.route} to {t.goesTo}, {t.departs} from {t.boardsAt}, Ksh{t.fareKsh},{" "}
              {t.seatsLeft === 0 ? "full" : `${t.seatsLeft} ${t.seatsLeft === 1 ? "seat" : "seats"} left`}{" "}
              <button type="button" disabled={t.seatsLeft === 0} onClick={() => onPick(t)}>
                Book the {t.route} at {t.departs}
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function TripPage({
  client,
  trip,
  user,
  onLogin,
  onBooked,
  onBack,
}: {
  client: Client;
  trip: Trip;
  user: User | null;
  onLogin: () => void;
  onBooked: (b: Booking) => void;
  onBack: () => void;
}) {
  const [seats, setSeats] = useState<SeatInfo | null>(null);
  const [seatsError, setSeatsError] = useState("");
  const [reload, setReload] = useState(0);
  const [seat, setSeat] = useState<number | null>(null);
  const [name, setName] = useState(user?.name ?? "");
  const [phone, setPhone] = useState(user?.phone ?? "");
  const [errors, setErrors] = useState<FieldErrors>({});
  const [alert, setAlert] = useState("");
  const [sending, setSending] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setSeatsError("");
    client.getSeats(trip.id).then(
      (s) => !cancelled && setSeats(s),
      (err: unknown) => !cancelled && setSeatsError(message(err)),
    );
    return () => {
      cancelled = true;
    };
  }, [client, trip.id, reload]);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setAlert("");
    const found = validateBooking({ name, phone, seat });
    setErrors(found);
    if (Object.keys(found).length > 0 || seat === null) return;
    setSending(true);
    try {
      const result = await client.book({ tripId: trip.id, seat, name: name.trim(), phone });
      if (result.kind === "booked") {
        onBooked(result.booking);
        return;
      }
      if (result.kind === "invalid") {
        setErrors(result.errors);
      } else {
        setAlert(`${result.message}. Pick another seat.`);
        setSeat(null);
        setReload((r) => r + 1); // someone beat us to it: fetch the seat map again
      }
    } catch (err) {
      if (err instanceof SessionExpiredError) return; // the App has already sent us to log in
      setAlert(`Booking failed: ${message(err)}`);
    } finally {
      setSending(false);
    }
  }

  function described(field: "name" | "phone") {
    return errors[field] ? { "aria-invalid": true, "aria-describedby": `${field}-error` } : {};
  }

  return (
    <section>
      <PageHeading>{title(trip)}</PageHeading>
      <button type="button" onClick={onBack}>
        Back to trips
      </button>
      {seatsError && <p role="alert">Couldn't load seats: {seatsError}</p>}
      {!seatsError && seats === null && <p role="status">Loading seats...</p>}
      <form onSubmit={handleSubmit} noValidate>
        {seats && (
          <fieldset aria-describedby={errors.seat ? "seat-error" : undefined}>
            <legend>Choose a seat</legend>
            {Array.from({ length: seats.seatCount }, (_, i) => i + 1).map((n) => {
              const taken = seats.taken.includes(n);
              return (
                <label key={n}>
                  <input
                    type="radio"
                    name="seat"
                    value={n}
                    checked={seat === n}
                    disabled={taken}
                    onChange={() => setSeat(n)}
                  />
                  Seat {n}
                  {taken && " (taken)"}
                </label>
              );
            })}
          </fieldset>
        )}
        {errors.seat && <p id="seat-error">{errors.seat}</p>}

        <label htmlFor="name">Name</label>
        <input id="name" autoComplete="name" value={name} onChange={(e) => setName(e.target.value)} {...described("name")} />
        {errors.name && <p id="name-error">{errors.name}</p>}

        <label htmlFor="phone">Phone</label>
        <input id="phone" type="tel" autoComplete="tel" value={phone} onChange={(e) => setPhone(e.target.value)} {...described("phone")} />
        {errors.phone && <p id="phone-error">{errors.phone}</p>}

        {user ? (
          <button type="submit" disabled={sending}>
            {sending ? "Booking..." : `Book seat, Ksh${trip.fareKsh}`}
          </button>
        ) : (
          <button type="button" onClick={onLogin}>
            Log in to book
          </button>
        )}
      </form>
      {alert && <p role="alert">{alert}</p>}
    </section>
  );
}

function LoginPage({
  client,
  notice,
  onLoggedIn,
  onCancel,
}: {
  client: Client;
  notice: string;
  onLoggedIn: (u: User) => void;
  onCancel: () => void;
}) {
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [sending, setSending] = useState(false);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");
    setSending(true);
    try {
      onLoggedIn(await client.login(phone, password));
    } catch (err) {
      setError(message(err));
      setPassword("");
      setSending(false);
    }
  }

  return (
    <section>
      <PageHeading>Log in</PageHeading>
      {notice && <p>{notice}</p>}
      <form onSubmit={handleSubmit}>
        <label htmlFor="login-phone">Phone</label>
        <input id="login-phone" type="tel" autoComplete="username" value={phone} onChange={(e) => setPhone(e.target.value)} />
        <label htmlFor="login-password">Password</label>
        <input
          id="login-password"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <button type="submit" disabled={sending}>
          Log in
        </button>
        <button type="button" onClick={onCancel}>
          Cancel
        </button>
      </form>
      {error && <p role="alert">{error}</p>}
    </section>
  );
}

function DonePage({ trip, booking, onAgain }: { trip: Trip; booking: Booking; onAgain: () => void }) {
  return (
    <section>
      <PageHeading>Booked!</PageHeading>
      <p>
        Reference {booking.reference}. Seat {booking.seat} on the {title(trip)}. Ksh{booking.totalKsh}. We'll SMS {booking.phone}.
      </p>
      <button type="button" onClick={onAgain}>
        Book another trip
      </button>
    </section>
  );
}

/** App is the whole booking app. The README lists what it must do. */
export function App({ client }: { client: Client }) {
  const [user, setUser] = useState<User | null>(null);
  const [view, setView] = useState<View>({ name: "trips" });

  const ended = useCallback(() => {
    setUser(null);
    setView((v) => ({ name: "login", back: v.name === "login" ? v.back : v, notice: ENDED }));
  }, []);

  useEffect(() => {
    client.onExpired = ended;
    let cancelled = false;
    client.restore().then(
      (u) => !cancelled && setUser(u),
      () => {},
    );
    return () => {
      cancelled = true;
    };
  }, [client, ended]);

  async function logOut() {
    await client.logout();
    setUser(null);
  }

  let page;
  switch (view.name) {
    case "trips":
      page = <TripsPage client={client} onPick={(trip) => setView({ name: "trip", trip })} />;
      break;
    case "trip":
      page = (
        <TripPage
          key={`${view.trip.id}-${user?.phone ?? ""}`}
          client={client}
          trip={view.trip}
          user={user}
          onLogin={() => setView({ name: "login", back: view, notice: "Log in to book a seat." })}
          onBooked={(booking) => setView({ name: "done", trip: view.trip, booking })}
          onBack={() => setView({ name: "trips" })}
        />
      );
      break;
    case "login":
      page = (
        <LoginPage
          client={client}
          notice={view.notice}
          onLoggedIn={(u) => {
            setUser(u);
            setView(view.back);
          }}
          onCancel={() => setView(view.back)}
        />
      );
      break;
    case "done":
      page = <DonePage trip={view.trip} booking={view.booking} onAgain={() => setView({ name: "trips" })} />;
      break;
  }

  return (
    <>
      <header>
        <h1>Twende</h1>
        {user ? (
          <p>
            Habari, {user.name}{" "}
            <button type="button" onClick={logOut}>
              Log out
            </button>
          </p>
        ) : (
          view.name !== "login" && (
            <button type="button" onClick={() => setView({ name: "login", back: view, notice: "" })}>
              Log in
            </button>
          )
        )}
      </header>
      <main>{page}</main>
    </>
  );
}
