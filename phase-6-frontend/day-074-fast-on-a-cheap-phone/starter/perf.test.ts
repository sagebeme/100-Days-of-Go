// @vitest-environment node
import { describe, expect, it } from "vitest";
import { checkBudget, formatBytes, pickWidth, srcSet } from "./perf";

describe("formatBytes", () => {
  const cases: [number, string][] = [
    [0, "0 B"],
    [999, "999 B"],
    [1000, "1.0 kB"],
    [1536, "1.5 kB"],
    [182_449, "182.4 kB"],
    [999_999, "1000.0 kB"],
    [2_300_000, "2.3 MB"],
  ];
  for (const [n, want] of cases) {
    it(`${n}`, () => {
      expect(formatBytes(n)).toBe(want);
    });
  }
});

describe("checkBudget", () => {
  const budget = { js: 170_000, css: 30_000 };
  it("is quiet within budget", () => {
    expect(checkBudget([{ name: "assets/index-a1.js", bytes: 60_000 }, { name: "assets/index-b2.css", bytes: 4_000 }], budget)).toEqual([]);
  });
  it("adds up every JS file, including lazy chunks", () => {
    const assets = [
      { name: "assets/index-a1.js", bytes: 150_000 },
      { name: "assets/SeatMap-c3.js", bytes: 32_449 },
      { name: "assets/logo-d4.png", bytes: 900_000 },
    ];
    expect(checkBudget(assets, budget)).toEqual(["JS is 182.4 kB, over the 170.0 kB budget"]);
  });
  it("reports JS before CSS", () => {
    const assets = [
      { name: "a.css", bytes: 31_000 },
      { name: "b.js", bytes: 171_000 },
    ];
    expect(checkBudget(assets, budget)).toEqual(["JS is 171.0 kB, over the 170.0 kB budget", "CSS is 31.0 kB, over the 30.0 kB budget"]);
  });
});

describe("pickWidth", () => {
  const widths = [960, 320, 640];
  const cases: { css: number; dpr: number; want: number }[] = [
    { css: 300, dpr: 1, want: 320 },
    { css: 320, dpr: 1, want: 320 },
    { css: 320, dpr: 2, want: 640 },
    { css: 360, dpr: 2.75, want: 960 },
    { css: 1200, dpr: 1, want: 960 },
  ];
  for (const c of cases) {
    it(`${c.css}px at ${c.dpr}x`, () => {
      expect(pickWidth(c.css, c.dpr, widths)).toBe(c.want);
    });
  }
});

describe("srcSet", () => {
  it("lists every width, smallest first", () => {
    expect(srcSet("matatu", [960, 320, 640])).toBe("/img/matatu-320.webp 320w, /img/matatu-640.webp 640w, /img/matatu-960.webp 960w");
  });
});
