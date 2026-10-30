// The client for the capstone's Go API (see api/main.go for every endpoint). The types and errors
// are the contract the tests use; keep them as they are.

export const API_URL: string = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export interface User {
  phone: string;
  name: string;
}

export interface Trip {
  id: string;
  route: string;
  goesTo: string;
  boardsAt: string;
  departs: string; // "07:45"
  fareKsh: number;
  seatsLeft: number;
}

export interface SeatInfo {
  tripId: string;
  seatCount: number;
  taken: number[];
}

export interface BookingInput {
  tripId: string;
  seat: number;
  name: string;
  phone: string;
}

export interface Booking {
  reference: string;
  tripId: string;
  seat: number;
  phone: string;
  totalKsh: number;
}

export type Field = "name" | "phone" | "seat";
export type FieldErrors = Partial<Record<Field, string>>;

/** BookingResult is what POST /api/bookings can mean: 201, 422 or 409. */
export type BookingResult =
  | { kind: "booked"; booking: Booking }
  | { kind: "invalid"; errors: FieldErrors }
  | { kind: "conflict"; message: string };

/** ApiError is any other non-2xx answer. message is the body's "error", or "HTTP STATUS". */
export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export class LoginError extends Error {
  constructor() {
    super("Wrong phone number or password");
    this.name = "LoginError";
  }
}

export class SessionExpiredError extends Error {
  constructor() {
    super("Your session has ended. Please log in again.");
    this.name = "SessionExpiredError";
  }
}

/** Client is everything the App needs. Tests pass a fake one. */
export interface Client {
  restore(): Promise<User | null>;
  login(phone: string, password: string): Promise<User>;
  logout(): Promise<void>;
  listTrips(): Promise<Trip[]>;
  getSeats(tripId: string): Promise<SeatInfo>;
  book(input: BookingInput): Promise<BookingResult>;
  /** onExpired is called when a refresh fails during book, just before it throws SessionExpiredError. */
  onExpired: () => void;
}

/**
 * createClient talks to the Go API at baseUrl. Every request uses credentials: "include".
 * The requirements are in the README; the tests in api.test.ts spell out each one.
 */
export function createClient(baseUrl: string = API_URL): Client {
  // Your code here
  const notYet = () => Promise.reject(new Error("not written yet"));
  return {
    restore: () => Promise.resolve(null),
    login: notYet,
    logout: () => Promise.resolve(),
    listTrips: notYet,
    getSeats: notYet,
    book: notYet,
    onExpired: () => {},
  };
}
