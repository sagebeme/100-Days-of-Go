// @vitest-environment node
import { describe, expect, it } from "vitest";
import { MESSAGES, normalisePhone, validateBooking } from "./validate";

describe("normalisePhone", () => {
  const cases: [string, string | null][] = [
    ["0712345678", "254712345678"],
    ["0712 345 678", "254712345678"],
    ["+254110123456", "254110123456"],
    ["254-712-345-678", "254712345678"],
    ["0812345678", null],
    ["071234567", null],
    ["", null],
  ];
  for (const [raw, want] of cases) {
    it(JSON.stringify(raw), () => {
      expect(normalisePhone(raw)).toBe(want);
    });
  }
});

describe("validateBooking", () => {
  const good = { name: "Amina", phone: "0712345678", seat: 3 };
  const cases: { name: string; form: { name: string; phone: string; seat: number | null }; want: object }[] = [
    { name: "good", form: good, want: {} },
    { name: "blank name", form: { ...good, name: "  " }, want: { name: MESSAGES.nameMissing } },
    { name: "long name", form: { ...good, name: "x".repeat(51) }, want: { name: MESSAGES.nameLong } },
    { name: "bad phone", form: { ...good, phone: "12" }, want: { phone: MESSAGES.phone } },
    { name: "no seat", form: { ...good, seat: null }, want: { seat: MESSAGES.seat } },
    {
      name: "everything",
      form: { name: "", phone: "", seat: null },
      want: { name: MESSAGES.nameMissing, phone: MESSAGES.phone, seat: MESSAGES.seat },
    },
  ];
  for (const c of cases) {
    it(c.name, () => {
      expect(validateBooking(c.form)).toEqual(c.want);
    });
  }
});
