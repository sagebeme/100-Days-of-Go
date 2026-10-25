// Day 70: a login page, a protected page, logout, and what to do when the session ends.
import { type FormEvent, type ReactNode, useCallback, useEffect, useState } from "react";
import { type AuthApi, SessionExpiredError, type User } from "./auth";

/** RequireAuth shows children only when there is a user, and fallback otherwise. */
export function RequireAuth({ user, fallback, children }: { user: User | null; fallback: ReactNode; children: ReactNode }) {
  // Your code here
  return null;
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
  // Your code here
  return <main></main>;
}
