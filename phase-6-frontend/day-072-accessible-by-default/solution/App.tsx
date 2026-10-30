// Day 72: the booking form again, built so everyone can use it: keyboard, screen reader, low vision.
import { type FormEvent, type KeyboardEvent, useEffect, useRef, useState } from "react";

/**
 * SeatPicker is a <fieldset> with <legend>Choose a seat</legend> and one native radio per seat,
 * each with a <label>: "Seat 3", or "Seat 2 (taken)" and disabled when taken.
 */
export function SeatPicker({
  seatCount,
  taken,
  value,
  onChange,
  invalid = false,
}: {
  seatCount: number;
  taken: number[];
  value: number | null;
  onChange: (seat: number) => void;
  invalid?: boolean;
}) {
  const seats = Array.from({ length: seatCount }, (_, i) => i + 1);
  return (
    // Native radios give us arrow keys, Space, and "radio, 3 of 14" from screen readers for free.
    <fieldset id="seat" aria-describedby={invalid ? "seat-error" : undefined}>
      <legend>Choose a seat</legend>
      {seats.map((n) => {
        const isTaken = taken.includes(n);
        return (
          <label key={n}>
            <input
              type="radio"
              name="seat"
              value={n}
              checked={value === n}
              disabled={isTaken}
              onChange={() => onChange(n)}
            />
            Seat {n}
            {isTaken && " (taken)"}
          </label>
        );
      })}
    </fieldset>
  );
}

type Field = "name" | "seat";

/**
 * BookingForm has labelled Name and Phone inputs, a SeatPicker and a Book button.
 * - A submit with errors shows <div role="alert" tabIndex={-1}> with an <h2>There is a problem</h2>
 *   and a link per error ("Enter your name", "Choose a seat"), and moves focus to it. Clicking a
 *   link focuses that field. Each field with an error gets aria-invalid and aria-describedby.
 * - A good submit opens <div role="dialog" aria-modal="true"> named by its <h2>Confirm your
 *   booking</h2>, saying "Seat N for NAME", with Confirm and Cancel buttons. Focus moves to Confirm,
 *   Tab and Shift+Tab stay inside, and Escape or Cancel closes it and puts focus back on Book.
 * - Confirm closes it and shows <p role="status">Seat N is yours, NAME.</p>
 */
export function BookingForm({ seatCount, taken }: { seatCount: number; taken: number[] }) {
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [seat, setSeat] = useState<number | null>(null);
  const [errors, setErrors] = useState<Partial<Record<Field, string>>>({});
  const [confirming, setConfirming] = useState(false);
  const [done, setDone] = useState("");

  const summaryRef = useRef<HTMLDivElement>(null);
  const bookRef = useRef<HTMLButtonElement>(null);
  const confirmRef = useRef<HTMLButtonElement>(null);
  const cancelRef = useRef<HTMLButtonElement>(null);
  const [submits, setSubmits] = useState(0);

  // Focus moves after React has drawn the new elements, so it happens in effects.
  useEffect(() => {
    if (submits > 0 && Object.keys(errors).length > 0) summaryRef.current?.focus();
  }, [submits, errors]);
  useEffect(() => {
    if (confirming) confirmRef.current?.focus();
  }, [confirming]);

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const found: Partial<Record<Field, string>> = {};
    if (name.trim() === "") found.name = "Enter your name";
    if (seat === null) found.seat = "Choose a seat";
    setErrors(found);
    setSubmits((s) => s + 1);
    setDone("");
    if (Object.keys(found).length === 0) setConfirming(true);
  }

  function close() {
    setConfirming(false);
    bookRef.current?.focus(); // put focus back where the user was
  }

  function trapKeys(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === "Escape") {
      close();
    } else if (e.key === "Tab") {
      // Two buttons: Tab from Cancel goes to Confirm, Shift+Tab from Confirm goes to Cancel.
      const first = confirmRef.current;
      const last = cancelRef.current;
      if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first?.focus();
      } else if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last?.focus();
      }
    }
  }

  function focusField(field: Field) {
    // For the seat group, focus the first radio that can take focus.
    const el =
      field === "name" ? document.getElementById("name") : document.querySelector<HTMLInputElement>("#seat input:not(:disabled)");
    el?.focus();
  }

  const fields: Field[] = ["name", "seat"];
  return (
    <main>
      <h1>Book a seat</h1>
      <form onSubmit={handleSubmit} noValidate>
        {Object.keys(errors).length > 0 && (
          <div role="alert" tabIndex={-1} ref={summaryRef} aria-labelledby="summary-title" className="error-summary">
            <h2 id="summary-title">There is a problem</h2>
            <ul>
              {fields
                .filter((f) => errors[f])
                .map((f) => (
                  <li key={f}>
                    <a
                      href={`#${f}`}
                      onClick={(e) => {
                        e.preventDefault();
                        focusField(f);
                      }}
                    >
                      {errors[f]}
                    </a>
                  </li>
                ))}
            </ul>
          </div>
        )}

        <label htmlFor="name">Name</label>
        <input
          id="name"
          autoComplete="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          aria-invalid={errors.name ? true : undefined}
          aria-describedby={errors.name ? "name-error" : undefined}
        />
        {errors.name && <p id="name-error">{errors.name}</p>}

        <label htmlFor="phone">Phone</label>
        <input id="phone" type="tel" autoComplete="tel" value={phone} onChange={(e) => setPhone(e.target.value)} />

        <SeatPicker seatCount={seatCount} taken={taken} value={seat} onChange={setSeat} invalid={!!errors.seat} />
        {errors.seat && <p id="seat-error">{errors.seat}</p>}

        <button type="submit" ref={bookRef}>
          Book
        </button>
      </form>

      {confirming && (
        <div role="dialog" aria-modal="true" aria-labelledby="confirm-title" onKeyDown={trapKeys}>
          <h2 id="confirm-title">Confirm your booking</h2>
          <p>
            Seat {seat} for {name.trim()}
          </p>
          <button
            type="button"
            ref={confirmRef}
            onClick={() => {
              setDone(`Seat ${seat} is yours, ${name.trim()}.`);
              close();
            }}
          >
            Confirm
          </button>
          <button type="button" ref={cancelRef} onClick={close}>
            Cancel
          </button>
        </div>
      )}
      <p role="status">{done}</p>
    </main>
  );
}
