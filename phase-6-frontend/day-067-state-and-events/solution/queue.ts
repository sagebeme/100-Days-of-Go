// Pure functions for the queue. They never change the list they're given; they return a new one.

/** Passenger is someone waiting at the stage. id is unique even when two people share a name. */
export interface Passenger {
  id: number;
  name: string;
}

/** addPassenger returns a new list with the passenger added at the end. A blank name (after trimming) changes nothing. */
export function addPassenger(list: Passenger[], name: string, id: number): Passenger[] {
  const trimmed = name.trim();
  if (trimmed === "") {
    return list;
  }
  return [...list, { id, name: trimmed }];
}

/** removePassenger returns a new list without the passenger with that id. */
export function removePassenger(list: Passenger[], id: number): Passenger[] {
  return list.filter((p) => p.id !== id);
}
