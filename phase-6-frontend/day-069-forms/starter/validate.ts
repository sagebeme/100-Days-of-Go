// Client-side checks. They mirror api/main.go word for word, so the passenger sees the same
// message whether the browser or the server catches the mistake.

/** BookingInput is the form's data, as the server expects it. */
export interface BookingInput {
  name: string;
  phone: string;
  route: string;
  seats: number;
}

export type Field = keyof BookingInput; // "name" | "phone" | "route" | "seats"

/** FieldErrors has a message for each field that's wrong, and no key for fields that are fine. */
export type FieldErrors = Partial<Record<Field, string>>;

export const ROUTE_NUMBERS = ["46K", "111", "125", "45", "58"];

export const MESSAGES = {
  nameMissing: "Enter your name",
  nameLong: "Name is too long (50 characters at most)",
  phone: "Enter a Kenyan mobile number like 0712 345 678",
  route: "Choose a route",
  seats: "Book between 1 and 14 seats",
} as const;

/**
 * normalisePhone removes spaces and dashes, then turns 07XXXXXXXX, 01XXXXXXXX, 2547... or +2547...
 * into 254XXXXXXXXX. It returns null for anything else.
 */
export function normalisePhone(raw: string): string | null {
  // Your code here
  return null;
}

/**
 * validateBooking applies the same rules as Validate in api/main.go:
 * name trimmed not empty and at most 50 characters; phone accepted by normalisePhone;
 * route one of ROUTE_NUMBERS; seats a whole number from 1 to 14.
 */
export function validateBooking(input: BookingInput): FieldErrors {
  // Your code here
  return {};
}
