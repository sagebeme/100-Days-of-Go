// The landing page: welcome back, the kata judge replay, and the route of phases.
(function () {
  var days = window.GO_DAYS, phases = window.GO_PHASES, P = window.GO_PROGRESS;
  var state = P.load();
  var next = P.nextDay(state, days);
  var $ = function (id) { return document.getElementById(id); };

  // ----- welcome back -----
  if (next) {
    var d = days[next - 1];
    var done = days.filter(function (x) { return P.isDone(state, x.day); }).length;
    ["cta", "cta-2"].forEach(function (id) {
      $(id).href = P.dayUrl(next);
      $(id).textContent = "Continue with Day " + next + " →";
    });
    var streak = P.streak();
    $("resume").textContent = done === days.length
      ? "All 100 days done. Go build something people use."
      : done + " of 100 days done" + (streak > 1 ? " · " + streak + "-day streak" : "") + " · up next: Day " + next + ", " + d.title;
    $("resume").hidden = false;
    $("cta-text").textContent = "Up next: Day " + next + ", " + d.title + ". One hour.";
  }

  // ----- the judge replay: a recorded run of the Day 1 kata solution -----
  var out = $("judge-out");
  var lines = [
    '<span class="p">$</span> go run ./tools/kata -solution 1',
    '  <span class="ok">PASS</span>  case 1      3ms',
    '  <span class="ok">PASS</span>  case 2      2ms',
    '  <span class="ok">PASS</span>  case 3      2ms',
    '  <span class="ok">PASS</span>  case 4      2ms',
    '  <span class="ok">PASS</span>  case 5     15ms',
    '',
    '<span class="ok">All 5 cases passed in 24ms. Kata done.</span>'
  ];
  var running = false;
  $("run-judge").addEventListener("click", function () {
    if (running) return;
    var still = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (still) { out.innerHTML = lines.join("\n"); return; }
    running = true;
    out.innerHTML = "";
    lines.forEach(function (line, i) {
      setTimeout(function () {
        out.innerHTML += (i ? "\n" : "") + line;
        if (i === lines.length - 1) running = false;
      }, 140 + i * 170);
    });
  });

  // ----- the route -----
  var currentPhase = next ? days[next - 1].phase : 1;
  var list = $("timeline");
  phases.forEach(function (ph) {
    var own = days.filter(function (x) { return x.phase === ph.n; });
    var written = own.filter(function (x) { return x.path; }).length;
    var finished = own.filter(function (x) { return P.isDone(state, x.day); }).length;
    var pct = Math.round((finished / own.length) * 100);
    var isHere = ph.n === currentPhase;
    var li = document.createElement("li");
    li.className = "stage" + (isHere ? " here" : "");
    li.innerHTML =
      '<span class="stop" aria-hidden="true">' + ph.n + "</span>" +
      '<div class="card">' +
        "<small>Phase " + ph.n + " · Days " + ph.first + "–" + ph.last +
        (written < own.length ? " · " + written + " of " + own.length + " written" : "") + "</small>" +
        "<h3>" + P.esc(ph.name) + (isHere ? '<span class="here-tag">' + (next ? "You are here" : "Start here") + "</span>" : "") + "</h3>" +
        "<p><b>You learn:</b> " + P.esc(ph.learn) + ".</p>" +
        "<p><b>Ends with:</b> " + P.esc(ph.ends) + ".</p>" +
        '<div class="meter"><div class="progress" role="progressbar" aria-label="Phase ' + ph.n + ' progress" aria-valuemin="0" aria-valuemax="' + own.length + '" aria-valuenow="' + finished + '"><span style="width:' + pct + '%"></span></div>' + finished + "/" + own.length + " done</div>" +
        "<details" + (isHere ? " open" : "") + "><summary>Days " + ph.first + "–" + ph.last + '</summary><ul class="day-list"></ul></details>' +
      "</div>";
    var ul = li.querySelector(".day-list");
    own.forEach(function (x) {
      var p = P.percent(state, x.day);
      var status = p === 100 ? '<span class="label label-done">Done</span>'
        : x.day === next ? '<span class="label label-next">Up next</span>'
        : !x.path ? '<span class="label label-planned">Planned</span>'
        : x.capstone ? '<span class="label label-cap">Capstone</span>'
        : p > 0 ? '<span class="label">' + p + "%</span>" : "";
      var item = document.createElement("li");
      item.innerHTML = '<a href="' + P.dayUrl(x.day) + '"' + (x.path ? "" : ' class="planned"') + '><span class="n">Day ' + x.day + '</span><span class="t">' + P.esc(x.title) + "</span>" + status + "</a>";
      ul.appendChild(item);
    });
    list.appendChild(li);
  });
})();
