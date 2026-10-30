// Pretend this is heavy: a big drawing library, icons, animations. Most passengers never open it,
// so it shouldn't be in the first download. Finished: use as is. React.lazy needs a default export.
export default function SeatMap() {
  const seats = Array.from({ length: 14 }, (_, i) => i + 1);
  return (
    <section>
      <h2>Seat map</h2>
      <ol className="seat-map">
        {seats.map((n) => (
          <li key={n}>Seat {n}</li>
        ))}
      </ol>
    </section>
  );
}
