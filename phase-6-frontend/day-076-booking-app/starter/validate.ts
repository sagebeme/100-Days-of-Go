// The browser's copy of the server's booking rules (api/main.go). Same rules, same words.
import type { FieldErrors } from "./api";

export const MESSAGES = {
  nameMissing: "Enter your name",
  nameLong: "Name is too long (50 characters at most)",
  phone: "Enter a Kenyan mobile number like 0712 345 678",
  seat: "Choose a seat",
} as const;

/** normalisePhone: Day 69's rule. "0712 345 678" → "254712345678"; anything else → null. */
export function normalisePhone(raw: string): string | null {
  // Your code here
  return null;
}

/** validateBooking checks name (trimmed, 1 to 50 characters), phone (normalisePhone) and seat (chosen). */
export function validateBooking(form: { name: string; phone: string; seat: number | null }): FieldErrors {
  // Your code here
  return {};
}
