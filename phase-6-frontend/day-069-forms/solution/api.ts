// The booking API client. This file is finished: today's work is in validate.ts and App.tsx.
import type { BookingInput, FieldErrors } from "./validate";

export const API_URL: string = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

/** Booking is what the server sends back on 201 Created. */
export interface Booking {
  reference: string;
  phone: string;
  route: string;
  seats: number;
}

/** BookingResult is a booking, or the server's field errors from a 422. */
export type BookingResult = { ok: true; booking: Booking } | { ok: false; errors: FieldErrors };

/** createBooking POSTs the form as JSON. Any status other than 201 or 422 throws an Error. */
export async function createBooking(input: BookingInput): Promise<BookingResult> {
  const res = await fetch(`${API_URL}/api/bookings`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (res.status === 201) {
    return { ok: true, booking: (await res.json()) as Booking };
  }
  if (res.status === 422) {
    const body = (await res.json()) as { errors: FieldErrors };
    return { ok: false, errors: body.errors };
  }
  throw new Error(`Booking failed (HTTP ${res.status}). Please try again.`);
}
