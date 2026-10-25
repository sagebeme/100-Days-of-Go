// Day 67: state and events.
import { useState } from "react";
import { type Passenger, addPassenger, removePassenger } from "./queue";

/**
 * SeatCounter shows "N of capacity seats taken" with a Board and an Alight button.
 * Board is disabled when full, Alight when empty. When full it shows
 * <p role="status">Imejaa! The mat is full.</p>.
 */
export function SeatCounter({ capacity }: { capacity: number }) {
  // Your code here
  return <section></section>;
}

/**
 * StageQueue is the conductor's list of waiting passengers.
 * - a <form> with an input labelled "Passenger name" and a button "Add to queue"
 * - "In the queue: N"
 * - <ul aria-label="Queue"> with one <li> per passenger: <span className="name">, an input
 *   labelled "Stop for NAME (ID)", and a button "Board NAME" that removes them
 */
export function StageQueue() {
  // Your code here
  return <section></section>;
}
