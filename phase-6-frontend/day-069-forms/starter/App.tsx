// Day 69: a controlled form with validation on both sides.
import { type FormEvent, useState } from "react";
import { type Booking, createBooking } from "./api";
import { type Field, type FieldErrors, ROUTE_NUMBERS, validateBooking } from "./validate";

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
  // Your code here
  return <form></form>;
}
