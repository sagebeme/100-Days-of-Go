// Day 68: loading, error and success states for a request to Go.
import { useEffect, useState } from "react";
import { type Route, fetchRoutes } from "./api";

/**
 * RouteSearch has an input labelled "Destination". Whenever the text changes (and once on load)
 * it calls fetchRoutes, cancelling the previous request with an AbortController. It shows:
 * - while loading: <p role="status">Loading routes...</p>
 * - on error: <div role="alert">Couldn't load routes: MESSAGE <button>Try again</button></div>
 * - no results: "No routes match."
 * - results: <ul aria-label="Routes"> with "111 to Ngong, Ksh100" or "To Kibera, ask the conductor"
 */
export function RouteSearch() {
  // Your code here
  return <main></main>;
}
