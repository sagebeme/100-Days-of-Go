// The lesson page: renders a day's README.md straight from its folder, plus the checklist.
(function () {
  var days = window.GO_DAYS, phases = window.GO_PHASES, P = window.GO_PROGRESS;
  var $ = function (id) { return document.getElementById(id); };
  var params = new URLSearchParams(location.search);
  var n = parseInt(params.get("day"), 10);
  if (!(n >= 1 && n <= days.length)) n = 1;
  var day = days[n - 1];
  var phase = phases[day.phase - 1];
  var BLOB = P.REPO + "/blob/main/";
  var TREE = P.REPO + "/tree/main/";

  // ----- the header -----
  document.title = "Day " + n + ": " + day.title + " · 100 Days of Go";
  $("badge").textContent = n;
  $("title").textContent = day.title;
  if (day.topic) {
    $("topic").textContent = day.topic;
    $("topic").hidden = false;
  }
  $("crumb-phase").textContent = "Phase " + phase.n + ": " + phase.name;
  $("crumb-day").textContent = "Day " + n;
  var labels = ['<span class="label">Phase ' + phase.n + " · Day " + (n - phase.first + 1) + " of " + (phase.last - phase.first + 1) + "</span>"];
  if (day.capstone) labels.push('<span class="label label-cap">Capstone</span>');
  if (n >= 85) labels.push('<span class="label">No walkthrough</span>');
  if (!day.path) labels.push('<span class="label label-planned">Planned</span>');
  $("labels").innerHTML = labels.join("");

  var lesson = $("lesson");

  // ----- a day that isn't written yet -----
  if (!day.path) {
    $("check-panel").hidden = true;
    $("run-panel").hidden = true;
    var written = days.filter(function (d) { return d.path; });
    var latest = written[written.length - 1];
    lesson.innerHTML =
      '<div class="notice"><h2>This day is on the way</h2>' +
      "<p>Day " + n + " teaches <strong>" + P.esc(day.topic) + "</strong>. Days are written phase by phase, so its folder isn't in the repo yet.</p>" +
      "<p>" + (latest ? 'The newest written day is <a href="' + P.dayUrl(latest.day) + '">Day ' + latest.day + ": " + P.esc(latest.title) + "</a>. " : "") +
      'Watch the <a href="' + P.REPO + '">repo on GitHub</a> to hear when this one lands.</p></div>';
    pager();
    return;
  }

  $("cmd").textContent = "go test ./" + day.path + "/starter";
  if (day.kata) {
    $("cmd-kata").textContent = "go run ./tools/kata " + n;
    $("cmd-kata").hidden = false;
  }
  $("gh-folder").href = TREE + day.path;
  $("gh-solution").href = TREE + day.path + "/solution";

  // ----- the lesson -----
  fetch("../" + day.path + "/README.md")
    .then(function (r) {
      if (!r.ok) throw new Error(r.status);
      return r.text();
    })
    .then(function (md) {
      lesson.innerHTML = marked.parse(md);
      enhance();
    })
    .catch(function () {
      lesson.innerHTML =
        '<div class="notice"><h2>This lesson couldn\'t load</h2>' +
        "<p>Browsers won't let a page opened straight from your disk read other files. From the course folder, run " +
        "<code>python3 -m http.server 8000</code> and open <code>http://localhost:8000/site/</code>.</p>" +
        '<p>Or read it on GitHub: <a href="' + BLOB + day.path + '/README.md">' + P.esc(day.path) + "/README.md</a></p></div>";
    });

  // A link in a README is relative to its folder. Another day's README opens here; any other file
  // in the repo opens on GitHub.
  function rewrite(a) {
    var href = a.getAttribute("href");
    if (!href || /^(https?:|mailto:|#)/.test(href)) return;
    var resolved = new URL(href, "https://repo.invalid/" + day.path + "/");
    var path = decodeURIComponent(resolved.pathname.slice(1)).replace(/\/$/, "");
    var target = days.find(function (d) { return d.path && (path === d.path || path === d.path + "/README.md"); });
    if (target) {
      a.href = P.dayUrl(target.day) + resolved.hash;
      return;
    }
    a.href = (/\.[a-z0-9]+$/i.test(path) ? BLOB : TREE) + path + resolved.hash;
  }

  function enhance() {
    lesson.querySelectorAll("a[href]").forEach(rewrite);

    lesson.querySelectorAll("table").forEach(function (t) {
      var wrap = document.createElement("div");
      wrap.className = "table-wrap";
      t.parentNode.insertBefore(wrap, t);
      wrap.appendChild(t);
    });

    lesson.querySelectorAll("pre").forEach(function (pre) {
      var code = pre.querySelector("code");
      // Highlight only blocks that name their language: the rest are output and examples.
      if (code && /language-/.test(code.className) && window.hljs) {
        try { hljs.highlightElement(code); } catch (e) { /* unknown language: leave it plain */ }
      }
      var wrap = document.createElement("div");
      wrap.className = "code";
      pre.parentNode.insertBefore(wrap, pre);
      wrap.appendChild(pre);
      var button = document.createElement("button");
      button.type = "button";
      button.className = "copy";
      button.textContent = "Copy";
      button.addEventListener("click", function () {
        var done = function (text) {
          button.textContent = text;
          setTimeout(function () { button.textContent = "Copy"; }, 1600);
        };
        var select = function () {
          var range = document.createRange();
          range.selectNodeContents(pre);
          var sel = getSelection();
          sel.removeAllRanges();
          sel.addRange(range);
          done("Press Ctrl+C");
        };
        try {
          navigator.clipboard.writeText(pre.innerText).then(function () { done("Copied"); }, select);
        } catch (e) {
          select();
        }
      });
      wrap.appendChild(button);
    });
  }

  // ----- the checklist -----
  var state = P.load();
  var items = P.checklist(day);
  var saved = P.ticks(state, n);
  var list = $("checklist");
  items.forEach(function (html, i) {
    var li = document.createElement("li");
    li.innerHTML = '<label><input type="checkbox" id="step-' + i + '"' + (saved[i] ? " checked" : "") + "><span>" + html + "</span></label>";
    li.querySelector("input").addEventListener("change", update);
    list.appendChild(li);
  });

  function update() {
    state[String(n)] = Array.prototype.map.call(list.querySelectorAll("input"), function (b) { return b.checked; });
    P.save(state);
    P.markToday();
    meter();
  }
  function meter() {
    var pct = P.percent(state, n);
    $("bar").style.width = pct + "%";
    $("pct").textContent = pct + "%";
    $("meter").setAttribute("aria-valuenow", String(pct));
    var msg = $("done-msg");
    if (pct === 100) {
      var phaseDone = days.filter(function (d) { return d.phase === phase.n; }).every(function (d) { return P.isDone(state, d.day); });
      msg.innerHTML = phaseDone
        ? "Phase " + phase.n + " complete: " + P.esc(phase.name) + ". " + (n < days.length ? '<a href="' + P.dayUrl(n + 1) + '">Next stage →</a>' : "That's the whole route.")
        : "Day " + n + " done. " + (n < days.length ? '<a href="' + P.dayUrl(n + 1) + '">On to Day ' + (n + 1) + " →</a>" : "");
      msg.hidden = false;
    } else {
      msg.hidden = true;
    }
  }
  meter();
  pager();

  // ----- previous and next -----
  function pager() {
    var nav = $("pager");
    function link(target, cls, label) {
      var a = document.createElement("a");
      a.className = cls;
      a.href = P.dayUrl(target.day);
      a.innerHTML = "<small>" + label + "</small><strong>" + P.esc(target.title) + "</strong>";
      nav.appendChild(a);
    }
    if (n > 1) link(days[n - 2], "prev", "← Day " + (n - 1));
    if (n < days.length) link(days[n], "next", "Day " + (n + 1) + " →");
  }
})();
