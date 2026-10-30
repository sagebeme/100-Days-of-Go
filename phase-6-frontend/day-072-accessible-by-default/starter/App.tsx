// Day 72: the booking form again, built so everyone can use it: keyboard, screen reader, low vision.
import { type FormEvent, type KeyboardEvent, useEffect, useRef, useState } from "react";

/**
 * SeatPicker is a <fieldset> with <legend>Choose a seat</legend> and one native radio per seat,
 * each with a <label>: "Seat 3", or "Seat 2 (taken)" and disabled when taken.
 */
export function SeatPicker({
  seatCount,
  taken,
  value,
  onChange,
}: {
  seatCount: number;
  taken: number[];
  value: number | null;
  onChange: (seat: number) => void;
}) {
  // Your code here
  return <div></div>;
}

/**
 * BookingForm has labelled Name and Phone inputs, a SeatPicker and a Book button.
 * - A submit with errors shows <div role="alert" tabIndex={-1}> with an <h2>There is a problem</h2>
 *   and a link per error ("Enter your name", "Choose a seat"), and moves focus to it. Clicking a
 *   link focuses that field. Each field with an error gets aria-invalid and aria-describedby.
 * - A good submit opens <div role="dialog" aria-modal="true"> named by its <h2>Confirm your
 *   booking</h2>, saying "Seat N for NAME", with Confirm and Cancel buttons. Focus moves to Confirm,
 *   Tab and Shift+Tab stay inside, and Escape or Cancel closes it and puts focus back on Book.
 * - Confirm closes it and shows <p role="status">Seat N is yours, NAME.</p>
 */
export function BookingForm({ seatCount, taken }: { seatCount: number; taken: number[] }) {
  // Your code here
  return <form></form>;
}
