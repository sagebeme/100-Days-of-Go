// Day 69: a controlled form with validation on both sides.
import { type FormEvent, useState } from "react";
import { type Booking, createBooking } from "./api";
import { type Field, type FieldErrors, ROUTE_NUMBERS, validateBooking } from "./validate";

/** A form's inputs are always strings; seats becomes a number only when we send it. */
interface FormValues {
  name: string;
  phone: string;
  route: string;
  seats: string;
}

/**
 * BookingForm is a <form> with four labelled fields: Name, Phone, Route (a <select> with
 * "Choose..." then ROUTE_NUMBERS) and Seats (<input type="number">, starting at 1).
 * On submit it runs validateBooking; if that finds errors it shows them and sends nothing.
 * Otherwise it calls createBooking and shows the server's field errors, or
 * "Booked! Reference REF. We'll SMS PHONE.", or the thrown error in role="alert".
 * Each error is a <p id="FIELD-error">, and its input has aria-invalid and aria-describedby.
 * The submit button says "Book" and becomes "Booking..." (disabled) while sending.
 */
export function BookingForm() {
  const [values, setValues] = useState<FormValues>({ name: "", phone: "", route: "", seats: "1" });
  const [errors, setErrors] = useState<FieldErrors>({});
  const [sending, setSending] = useState(false);
  const [booking, setBooking] = useState<Booking | null>(null);
  const [failure, setFailure] = useState("");

  // One change handler for every field: the input's name says which value to update.
  function change(e: { target: { name: string; value: string } }) {
    setValues({ ...values, [e.target.name]: e.target.value });
  }

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setFailure("");
    setBooking(null);
    // An empty number box gives NaN, which validateBooking rejects.
    const input = { ...values, seats: values.seats === "" ? Number.NaN : Number(values.seats) };
    const found = validateBooking(input);
    setErrors(found);
    if (Object.keys(found).length > 0) return;

    setSending(true);
    try {
      const result = await createBooking(input);
      if (result.ok) {
        setBooking(result.booking);
      } else {
        setErrors(result.errors); // the server knows things we can't, like seats left
      }
    } catch (err) {
      setFailure(err instanceof Error ? err.message : String(err));
    } finally {
      setSending(false);
    }
  }

  // props for an input: wires up aria-invalid and aria-describedby when the field has an error.
  function errorProps(field: Field) {
    return errors[field] ? { "aria-invalid": true, "aria-describedby": `${field}-error` } : {};
  }
  function errorText(field: Field) {
    return errors[field] ? (
      <p id={`${field}-error`} className="error">
        {errors[field]}
      </p>
    ) : null;
  }

  return (
    <form onSubmit={handleSubmit} noValidate>
      <h1>Book a seat</h1>
      <label htmlFor="name">Name</label>
      <input id="name" name="name" autoComplete="name" value={values.name} onChange={change} {...errorProps("name")} />
      {errorText("name")}

      <label htmlFor="phone">Phone</label>
      <input id="phone" name="phone" type="tel" autoComplete="tel" value={values.phone} onChange={change} {...errorProps("phone")} />
      {errorText("phone")}

      <label htmlFor="route">Route</label>
      <select id="route" name="route" value={values.route} onChange={change} {...errorProps("route")}>
        <option value="">Choose...</option>
        {ROUTE_NUMBERS.map((n) => (
          <option key={n} value={n}>
            {n}
          </option>
        ))}
      </select>
      {errorText("route")}

      <label htmlFor="seats">Seats</label>
      <input id="seats" name="seats" type="number" min={1} max={14} value={values.seats} onChange={change} {...errorProps("seats")} />
      {errorText("seats")}

      <button type="submit" disabled={sending}>
        {sending ? "Booking..." : "Book"}
      </button>
      {booking && (
        <p role="status">
          Booked! Reference {booking.reference}. We'll SMS {booking.phone}.
        </p>
      )}
      {failure && <p role="alert">{failure}</p>}
    </form>
  );
}
