// @vitest-environment node
import { describe, expect, it } from "vitest";
import { type BookingInput, MESSAGES, normalisePhone, validateBooking } from "./validate";

describe("normalisePhone", () => {
  const cases: { raw: string; want: string | null }[] = [
    { raw: "0712345678", want: "254712345678" },
    { raw: "0712 345 678", want: "254712345678" },
    { raw: "0110-123-456", want: "254110123456" },
    { raw: "+254712345678", want: "254712345678" },
    { raw: "254712345678", want: "254712345678" },
    { raw: "0812345678", want: null },
    { raw: "071234567", want: null },
    { raw: "07123456789", want: null },
    { raw: "+255712345678", want: null },
    { raw: "", want: null },
  ];
  for (const c of cases) {
    it(JSON.stringify(c.raw), () => {
      expect(normalisePhone(c.raw), `normalisePhone(${JSON.stringify(c.raw)})`).toBe(c.want);
    });
  }
});

describe("validateBooking", () => {
  const good: BookingInput = { name: "Amina Wanjiru", phone: "0712 345 678", route: "58", seats: 2 };
  const cases: { name: string; input: BookingInput; want: object }[] = [
    { name: "a good booking", input: good, want: {} },
    { name: "blank name", input: { ...good, name: "   " }, want: { name: MESSAGES.nameMissing } },
    { name: "51-character name", input: { ...good, name: "a".repeat(51) }, want: { name: MESSAGES.nameLong } },
    { name: "50 emoji count as 50", input: { ...good, name: "😀".repeat(50) }, want: {} },
    { name: "bad phone", input: { ...good, phone: "12345" }, want: { phone: MESSAGES.phone } },
    { name: "no route", input: { ...good, route: "" }, want: { route: MESSAGES.route } },
    { name: "zero seats", input: { ...good, seats: 0 }, want: { seats: MESSAGES.seats } },
    { name: "15 seats", input: { ...good, seats: 15 }, want: { seats: MESSAGES.seats } },
    { name: "half a seat", input: { ...good, seats: 1.5 }, want: { seats: MESSAGES.seats } },
    { name: "an empty number box", input: { ...good, seats: Number.NaN }, want: { seats: MESSAGES.seats } },
    {
      name: "everything wrong",
      input: { name: "", phone: "", route: "", seats: 0 },
      want: { name: MESSAGES.nameMissing, phone: MESSAGES.phone, route: MESSAGES.route, seats: MESSAGES.seats },
    },
  ];
  for (const c of cases) {
    it(c.name, () => {
      expect(validateBooking(c.input)).toEqual(c.want);
    });
  }
});
