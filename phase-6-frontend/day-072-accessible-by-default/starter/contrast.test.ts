// @vitest-environment node
import { describe, expect, it } from "vitest";
import { contrastRatio, meetsAA, relativeLuminance } from "./contrast";
import { THEME } from "./theme";

describe("relativeLuminance", () => {
  const cases: { hex: string; want: number }[] = [
    { hex: "#000000", want: 0 },
    { hex: "#ffffff", want: 1 },
    { hex: "#FF0000", want: 0.2126 },
    { hex: "#808080", want: 0.2158605 },
  ];
  for (const c of cases) {
    it(c.hex, () => {
      expect(relativeLuminance(c.hex)).toBeCloseTo(c.want, 5);
    });
  }
});

describe("contrastRatio", () => {
  const cases: { a: string; b: string; want: number }[] = [
    { a: "#000000", b: "#ffffff", want: 21 },
    { a: "#ffffff", b: "#000000", want: 21 },
    { a: "#777777", b: "#777777", want: 1 },
    { a: "#767676", b: "#ffffff", want: 4.54 },
    { a: "#777777", b: "#ffffff", want: 4.48 },
  ];
  for (const c of cases) {
    it(`${c.a} on ${c.b}`, () => {
      expect(contrastRatio(c.a, c.b)).toBeCloseTo(c.want, 2);
    });
  }
});

describe("meetsAA", () => {
  it("uses 4.5 for normal text and 3 for large text", () => {
    expect([meetsAA(4.5, false), meetsAA(4.49, false), meetsAA(3, true), meetsAA(2.99, true)]).toEqual([true, false, true, false]);
  });
});

describe("THEME", () => {
  const pairs: [string, string, string][] = [
    ["text", THEME.text, THEME.background],
    ["muted", THEME.muted, THEME.background],
    ["error", THEME.error, THEME.background],
    ["button", THEME.buttonText, THEME.buttonBackground],
  ];
  for (const [name, fg, bg] of pairs) {
    it(`${name} text is readable`, () => {
      const ratio = contrastRatio(fg, bg);
      expect(meetsAA(ratio, false), `${name}: ${fg} on ${bg} is ${ratio.toFixed(2)}:1`).toBe(true);
    });
  }
});
