// Day 65: TypeScript for Go developers. Fill in each function so the tests pass.

/** Route is one matatu route. Like a Go struct, but it's only a shape: no methods, no constructor. */
export interface Route {
  number: string;
  operator: string;
  boardsAt: string;
  goesTo: string;
  fareKsh: number;
}

/** ROUTES are real Nairobi routes, checked on nishukishe.com in October 2026. */
export const ROUTES: Route[] = [
  { number: "46K", operator: "Citi Hoppa", boardsAt: "Kencom / Ambassadeur", goesTo: "Kawangware", fareKsh: 100 },
  { number: "111", operator: "MCA Sacco", boardsAt: "Railways", goesTo: "Ngong", fareKsh: 100 },
  { number: "125", operator: "City Shuttle", boardsAt: "Railways", goesTo: "Ongata Rongai", fareKsh: 100 },
  { number: "45", operator: "Zamzam", boardsAt: "Ronald Ngala Street", goesTo: "Githurai", fareKsh: 100 },
  { number: "58", operator: "Buruburu 58 Travellers Sacco", boardsAt: "Ambassadeur", goesTo: "Buruburu", fareKsh: 150 },
];

/** Payment is a union: exactly one of these shapes, told apart by `kind`. */
export type Payment =
  | { kind: "mpesa"; phone: string; amountKsh: number }
  | { kind: "cash"; amountKsh: number }
  | { kind: "card"; last4: string; amountKsh: number };

/** describePayment says how a fare was paid, e.g. "Ksh100 by M-Pesa from 0712345678". */
export function describePayment(p: Payment): string {
  // Your code here
  return "";
}

/** fareFor returns the fare for a route number, or undefined if there's no such route. */
export function fareFor(routes: Route[], number: string): number | undefined {
  // Your code here
  return undefined;
}

/** groupBy puts items into lists by key, keeping the order keys were first seen. */
export function groupBy<T, K>(items: T[], key: (item: T) => K): Map<K, T[]> {
  // Your code here
  return new Map();
}

/** parseRoute checks that a value from JSON.parse really is a Route. It throws Error("invalid route: ...") if not. */
export function parseRoute(value: unknown): Route {
  // Your code here
  return { number: "", operator: "", boardsAt: "", goesTo: "", fareKsh: 0 };
}

/** totalFare looks up every route's fare at the same time and adds them. If any lookup fails, it fails. */
export async function totalFare(numbers: string[], lookup: (number: string) => Promise<number>): Promise<number> {
  // Your code here
  return 0;
}
