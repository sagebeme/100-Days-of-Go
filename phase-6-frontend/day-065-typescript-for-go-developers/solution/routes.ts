// Day 65: TypeScript for Go developers.

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
  switch (p.kind) {
    case "mpesa":
      return `Ksh${p.amountKsh} by M-Pesa from ${p.phone}`; // here TypeScript knows p has phone
    case "cash":
      return `Ksh${p.amountKsh} in cash`;
    case "card":
      return `Ksh${p.amountKsh} by card ending ${p.last4}`;
    default: {
      // If someone adds a fourth kind and forgets it here, this line stops compiling.
      const never: never = p;
      return never;
    }
  }
}

/** fareFor returns the fare for a route number, or undefined if there's no such route. */
export function fareFor(routes: Route[], number: string): number | undefined {
  return routes.find((r) => r.number === number)?.fareKsh;
}

/** groupBy puts items into lists by key, keeping the order keys were first seen. */
export function groupBy<T, K>(items: T[], key: (item: T) => K): Map<K, T[]> {
  const groups = new Map<K, T[]>();
  for (const item of items) {
    const k = key(item);
    const list = groups.get(k);
    if (list) {
      list.push(item);
    } else {
      groups.set(k, [item]);
    }
  }
  return groups;
}

/** parseRoute checks that a value from JSON.parse really is a Route. It throws Error("invalid route: ...") if not. */
export function parseRoute(value: unknown): Route {
  if (typeof value !== "object" || value === null) {
    throw new Error("invalid route: not an object");
  }
  const v = value as Record<string, unknown>;
  for (const field of ["number", "operator", "boardsAt", "goesTo"]) {
    if (typeof v[field] !== "string") {
      throw new Error(`invalid route: ${field} must be a string`);
    }
  }
  if (typeof v.fareKsh !== "number") {
    throw new Error("invalid route: fareKsh must be a number");
  }
  return {
    number: v.number as string,
    operator: v.operator as string,
    boardsAt: v.boardsAt as string,
    goesTo: v.goesTo as string,
    fareKsh: v.fareKsh,
  };
}

/** totalFare looks up every route's fare at the same time and adds them. If any lookup fails, it fails. */
export async function totalFare(numbers: string[], lookup: (number: string) => Promise<number>): Promise<number> {
  const fares = await Promise.all(numbers.map(lookup));
  return fares.reduce((sum, f) => sum + f, 0);
}
