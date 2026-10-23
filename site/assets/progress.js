// Progress, kept in this browser only: { "6": [true, true, false, false, false] } is Day 6's checklist.
// Shared by every page. No accounts, no server.
(function () {
  var KEY = "go100-progress";
  var REPO = "https://github.com/sagebeme/100-Days-of-Go";

  // The same five steps every day: the project, then the kata.
  function checklist(day) {
    var steps = [
      "Read the brief",
      "Wrote my code in <code>starter/</code>",
      "Tests pass: <code>go test ./" + day.path + "/starter</code>"
    ];
    steps.push(day.kata
      ? "Kata judged: <code>go run ./tools/kata " + day.day + "</code>"
      : "Explained today's idea out loud, in two sentences");
    steps.push("Compared with <code>solution/</code> and committed");
    return steps;
  }
  var STEPS = 5;

  function load() {
    try {
      var state = JSON.parse(localStorage.getItem(KEY) || "{}");
      return state && typeof state === "object" ? state : {};
    } catch (e) {
      return {};
    }
  }
  function save(state) {
    try { localStorage.setItem(KEY, JSON.stringify(state)); } catch (e) { /* private window: progress just isn't kept */ }
  }
  function ticks(state, day) {
    var saved = state[String(day)];
    return Array.isArray(saved) ? saved : [];
  }
  // 0 to 100 for one day.
  function percent(state, day) {
    var done = ticks(state, day).filter(Boolean).length;
    return Math.round((done / STEPS) * 100);
  }
  function isDone(state, day) { return percent(state, day) === 100; }

  // Where to go next: the first unfinished day after the last one touched. Null if nothing's touched.
  function nextDay(state, days) {
    var touched = days.filter(function (d) { return ticks(state, d.day).some(Boolean); });
    if (!touched.length) return null;
    var last = touched[touched.length - 1].day;
    if (!isDone(state, last)) return last;
    for (var n = last + 1; n <= days.length; n++) if (!isDone(state, n)) return n;
    for (n = 1; n <= days.length; n++) if (!isDone(state, n)) return n;
    return days.length;
  }

  // Days in a row with at least one tick, ending today or yesterday. Kept as a list of dates.
  var STREAK = "go100-streak";
  function markToday() {
    try {
      var dates = JSON.parse(localStorage.getItem(STREAK) || "[]");
      var today = new Date().toISOString().slice(0, 10);
      if (dates.indexOf(today) === -1) dates.push(today);
      localStorage.setItem(STREAK, JSON.stringify(dates.slice(-400)));
    } catch (e) { /* no storage: no streak */ }
  }
  function streak() {
    try {
      var dates = JSON.parse(localStorage.getItem(STREAK) || "[]");
      var set = {};
      dates.forEach(function (d) { set[d] = true; });
      var day = new Date();
      var key = function (d) { return d.toISOString().slice(0, 10); };
      if (!set[key(day)]) day.setUTCDate(day.getUTCDate() - 1);
      var count = 0;
      while (set[key(day)]) { count++; day.setUTCDate(day.getUTCDate() - 1); }
      return count;
    } catch (e) {
      return 0;
    }
  }

  function dayUrl(n) { return "day.html?day=" + n; }
  function pad(n) { return String(n).padStart(3, "0"); }

  // Escape text before it goes into innerHTML.
  function esc(text) {
    return String(text).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  }

  function backToTop() {
    var a = document.querySelector(".to-top");
    if (!a) return;
    var update = function () { a.classList.toggle("show", window.scrollY > 700); };
    window.addEventListener("scroll", update, { passive: true });
    update();
  }
  document.addEventListener("DOMContentLoaded", backToTop);

  window.GO_PROGRESS = {
    REPO: REPO, checklist: checklist, load: load, save: save, ticks: ticks, percent: percent,
    isDone: isDone, nextDay: nextDay, markToday: markToday, streak: streak, dayUrl: dayUrl, pad: pad, esc: esc
  };
})();
