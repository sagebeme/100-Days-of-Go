// @vitest-environment node
import { describe, expect, it } from "vitest";
import { type Passenger, addPassenger, removePassenger } from "./queue";

describe("addPassenger", () => {
  const cases: { name: string; list: Passenger[]; add: string; want: Passenger[] }[] = [
    { name: "to an empty queue", list: [], add: "Amina", want: [{ id: 1, name: "Amina" }] },
    {
      name: "at the end",
      list: [{ id: 1, name: "Amina" }],
      add: "Baraka",
      want: [{ id: 1, name: "Amina" }, { id: 2, name: "Baraka" }],
    },
    { name: "trims spaces", list: [], add: "  Chebet ", want: [{ id: 1, name: "Chebet" }] },
    { name: "ignores a blank name", list: [], add: "   ", want: [] },
  ];
  for (const c of cases) {
    it(c.name, () => {
      const before = structuredClone(c.list);
      const got = addPassenger(c.list, c.add, c.list.length + 1);
      expect(got, `addPassenger(..., ${JSON.stringify(c.add)})`).toEqual(c.want);
      expect(c.list, "the original list must not change").toEqual(before);
    });
  }
  it("returns a new array, so React sees the change", () => {
    const list: Passenger[] = [];
    expect(addPassenger(list, "Amina", 1)).not.toBe(list);
  });
});

describe("removePassenger", () => {
  it("removes only the passenger with that id, even if another shares the name", () => {
    const list: Passenger[] = [
      { id: 1, name: "Kamau" },
      { id: 2, name: "Kamau" },
      { id: 3, name: "Akinyi" },
    ];
    const got = removePassenger(list, 2);
    expect(got).toEqual([{ id: 1, name: "Kamau" }, { id: 3, name: "Akinyi" }]);
    expect(list.length, "the original list must not change").toBe(3);
  });
  it("leaves the list alone for an unknown id", () => {
    expect(removePassenger([{ id: 1, name: "Kamau" }], 9)).toEqual([{ id: 1, name: "Kamau" }]);
  });
});
