// Day 67: state and events.
import { type FormEvent, useState } from "react";
import { type Passenger, addPassenger, removePassenger } from "./queue";

/**
 * SeatCounter shows "N of capacity seats taken" with a Board and an Alight button.
 * Board is disabled when full, Alight when empty. When full it shows
 * <p role="status">Imejaa! The mat is full.</p>.
 */
export function SeatCounter({ capacity }: { capacity: number }) {
  const [taken, setTaken] = useState(0);
  const full = taken >= capacity;
  return (
    <section>
      <h2>Seats</h2>
      <p>
        {taken} of {capacity} seats taken
      </p>
      <button type="button" disabled={full} onClick={() => setTaken((t) => t + 1)}>
        Board
      </button>
      <button type="button" disabled={taken === 0} onClick={() => setTaken((t) => t - 1)}>
        Alight
      </button>
      {full && <p role="status">Imejaa! The mat is full.</p>}
    </section>
  );
}

/**
 * StageQueue is the conductor's list of waiting passengers.
 * - a <form> with an input labelled "Passenger name" and a button "Add to queue"
 * - "In the queue: N"
 * - <ul aria-label="Queue"> with one <li> per passenger: <span className="name">, an input
 *   labelled "Stop for NAME (ID)", and a button "Board NAME" that removes them
 */
export function StageQueue() {
  const [queue, setQueue] = useState<Passenger[]>([]);
  const [name, setName] = useState("");
  const [nextId, setNextId] = useState(1);

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); // stop the browser reloading the page
    if (name.trim() !== "") {
      setQueue(addPassenger(queue, name, nextId));
      setNextId(nextId + 1);
    }
    setName("");
  }

  return (
    <section>
      <h2>Queue</h2>
      <form onSubmit={handleSubmit}>
        <label>
          Passenger name <input value={name} onChange={(e) => setName(e.target.value)} />
        </label>
        <button type="submit">Add to queue</button>
      </form>
      <p>In the queue: {queue.length}</p>
      <ul aria-label="Queue">
        {queue.map((p) => (
          // The key is the id, never the index: it keeps each <li> (and its typed stop) with its passenger.
          <li key={p.id}>
            <span className="name">{p.name}</span>{" "}
            <label>
              Stop for {p.name} ({p.id}) <input />
            </label>{" "}
            <button type="button" onClick={() => setQueue(removePassenger(queue, p.id))}>
              Board {p.name}
            </button>
          </li>
        ))}
      </ul>
    </section>
  );
}
