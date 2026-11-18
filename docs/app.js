const IPA = "https://www.ipa.go.jp/shiken/mondai-kaiotu/";
const EXAMS = {
  r7a: ["令和7年度秋期", "nl10bi0000009lh8-att/2025r07a"],
  r7h: ["令和7年度春期", "nl10bi0000009lh8-att/2025r07h"],
  r6a: ["令和6年度秋期", "m42obm000000afqx-att/2024r06a"],
  r6h: ["令和6年度春期", "m42obm000000afqx-att/2024r06h"],
  r5a: ["令和5年度秋期", "ps6vr70000010d6y-att/2023r05a"],
  r5h: ["令和5年度春期", "ps6vr70000010d6y-att/2023r05h"],
  r4a: ["令和4年度秋期", "gmcbt80000008smf-att/2022r04a"],
  r4h: ["令和4年度春期", "gmcbt80000009sgk-att/2022r04h"],
};
const TUTORIAL = "https://ijuken.com/engine/api/public/demo_exam_start/02f50984f6a37331f9af54a7b401e41f/";
const GUIDE = "https://cbt-s.com/examinee/files/examination/17907371471.pdf";

const FIELD_NAMES = {
  db: "問6 データベース",
  nw: "問5 ネットワーク",
  pg: "問3 プログラミング",
  st: "問2 経営戦略",
};

const SOLVE = "25分で解き、20分で復習する。解答はPCで打つ。";
const q = (id, exam, label, detail = SOLVE) => ({ id, exam, title: `${EXAMS[exam][0]} ${label}`, min: 45, detail });
const task = (id, title, min, detail, extra = {}) => ({ id, title, min, detail, ...extra });

const DAYS = [
  {
    date: "2026-11-19", dow: "木", phase: "診断", title: "操作に慣れて、現在地を知る",
    goal: "CBTの画面を一度触り、主力3分野の実力を測ります。",
    finish: "操作に迷わない。3問の得点を記録できた。",
    point: "点が低くても問題ありません。今日は弱点を見つける日です。",
    tasks: [
      task("d19-tutorial", "公式チュートリアルで操作確認", 15, "問題選択、「選択」ボタン、メモ、マーカー、電卓を触る。PCで開く。", { link: [TUTORIAL, "チュートリアルを開く"], link2: [GUIDE, "操作説明PDF"] }),
      q("d19-q1", "r5a", "問1 情報セキュリティ"),
      q("d19-q9", "r5a", "問9 プロジェクトマネジメント"),
      q("d19-q11", "r5a", "問11 システム監査"),
    ],
  },
  {
    date: "2026-11-20", dow: "金", phase: "診断", title: "解く5分野を決める",
    goal: "残りの候補を試し、本番で解く分野と予備を確定します。",
    finish: "技術枠と予備を選択できた。",
    point: "迷ったら、得点の高さより「25分で解き切れたか」で選びます。",
    tasks: [
      q("d20-q10", "r5a", "問10 サービスマネジメント"),
      q("d20-tech1", "r5a", "技術候補1問目", "問3・問5・問6から得意そうな分野を解く。25分＋復習20分。"),
      q("d20-tech2", "r5a", "技術候補2問目", "問3・問5・問6から別の分野を解く。25分＋復習20分。"),
      task("d20-pick", "技術枠と予備を確定する", 15, "下で選ぶと、明日以降の課題に反映されます。", { picker: true }),
    ],
  },
  {
    date: "2026-11-21", dow: "土", phase: "分野別演習", title: "必須と主力3分野を固める",
    goal: "問1と主力3分野を2年度分、計8問解きます。",
    finish: "8問すべての間違いを原因別に分類できた。",
    point: "間違いは「知識不足・読み落とし・表現ミス」に分類します。",
    tasks: [
      q("d21-a1", "r6h", "問1 情報セキュリティ"), q("d21-a9", "r6h", "問9 プロジェクトマネジメント"),
      q("d21-a10", "r6h", "問10 サービスマネジメント"), q("d21-a11", "r6h", "問11 システム監査"),
      q("d21-b1", "r5h", "問1 情報セキュリティ"), q("d21-b9", "r5h", "問9 プロジェクトマネジメント"),
      q("d21-b10", "r5h", "問10 サービスマネジメント"), q("d21-b11", "r5h", "問11 システム監査"),
    ],
  },
  {
    date: "2026-11-22", dow: "日", phase: "分野別演習", title: "技術枠を仕上げ、弱点をまとめる",
    goal: "技術枠と予備を演習し、弱点をノート1枚に集約します。",
    finish: "本番直前に見返す弱点ノートが1枚できた。",
    point: "ノートは1枚に収めます。増やすほど見返せなくなります。",
    tasks: [
      q("d22-a-tech", "r6h", "{tech}"), q("d22-a-res", "r6h", "{reserve}"),
      q("d22-b-tech", "r5h", "{tech}"), q("d22-b-res", "r5h", "{reserve}"),
      task("d22-redo", "昨日間違えた2問を解き直す", 90, "解答の根拠を本文から説明できるまで繰り返す。"),
      task("d22-note", "弱点ノートを1枚にまとめる", 90, "用語、間違えた理由、次に気をつけることを書く。"),
    ],
  },
  {
    date: "2026-11-23", dow: "月祝", phase: "模試と補強", title: "150分模試で実力を測る", mock: "m1",
    goal: "本番と同じ150分で5問を解き、60点との差を把握します。",
    finish: "得点を記録し、不足分の原因が分かった。",
    point: "1問25分を超えたら、途中でも次へ進みます。",
    tasks: [
      task("d23-mock", "令和6年度秋期 150分模試", 150, "問1＋選んだ4問を通しで解く。途中で解答を見ない。", { exam: "r6a" }),
      task("d23-review", "模試の採点と復習", 150, "解答例と採点講評を読み、失点の原因を分類する。", { exam: "r6a" }),
      task("d23-fix", "弱点を補強する", 60, "失点の大きい分野の知識を確認し、ノートに追記する。"),
    ],
  },
  {
    date: "2026-11-24", dow: "火", phase: "模試と補強", title: "模試の失点を取り返す",
    goal: "模試で失点が大きかった問題を解き直します。",
    finish: "同じ原因で間違えなくなった。",
    point: "新しい問題より、間違えた問題の解き直しを優先します。",
    tasks: [
      task("d24-a", "失点が最も大きい問題を解き直す", 45, "解答の根拠となる本文に印を付ける。", { exam: "r6a" }),
      task("d24-b", "失点が2番目に大きい問題を解き直す", 45, "解答の根拠となる本文に印を付ける。", { exam: "r6a" }),
      task("d24-c", "失点が3番目に大きい問題を解き直す", 45, "解答の根拠となる本文に印を付ける。", { exam: "r6a" }),
      task("d24-note", "弱点ノートを更新する", 15, "今日分かったことを1行ずつ追記する。"),
    ],
  },
  {
    date: "2026-11-25", dow: "水", phase: "模試と補強", title: "必須の問1を安定させる",
    goal: "問1を25分で解き切れるようにし、最も弱い分野も補強します。",
    finish: "問1を25分以内に解き切れた。",
    point: "問1は必ず解く問題です。ここでの失点を減らします。",
    tasks: [
      q("d25-a", "r4a", "問1 情報セキュリティ"),
      q("d25-b", "r4h", "問1 情報セキュリティ"),
      q("d25-c", "r4a", "最も弱い分野", "これまでの結果で最も得点が低い分野を解く。25分＋復習20分。"),
      task("d25-note", "弱点ノートを更新する", 15, "セキュリティ用語の覚え漏れを追記する。"),
    ],
  },
  {
    date: "2026-11-26", dow: "木", phase: "模試と補強", title: "2回目の150分模試", mock: "m2",
    goal: "時間内に5問を解き切ることを最優先にします。",
    finish: "150分で5問すべてに解答できた。",
    point: "今日は解くだけで大丈夫です。復習は明日に回します。",
    tasks: [
      task("d26-mock", "令和7年度春期 150分模試", 150, "残り時間の目安を見ながら通しで解く。", { exam: "r7h" }),
    ],
  },
  {
    date: "2026-11-27", dow: "金", phase: "模試と補強", title: "模試を復習し、弱点を潰す",
    goal: "2回目の模試の失点を、原因から修正します。",
    finish: "間違えた問題の解答根拠を説明できる。",
    point: "1回目の模試と同じ間違いがないか確認します。",
    tasks: [
      task("d27-review", "模試の採点と復習", 90, "解答例と採点講評で失点の原因を確認する。", { exam: "r7h" }),
      task("d27-redo", "失点が最も大きい問題を解き直す", 45, "時間を計り、PCで解答を打つ。", { exam: "r7h" }),
      task("d27-note", "弱点ノートを更新する", 15, "本番前に見返す内容だけを残す。"),
    ],
  },
  {
    date: "2026-11-28", dow: "土", phase: "仕上げ", title: "本番と同じ時刻で最終模試", mock: "m3",
    goal: "17:00開始のリハーサルで、本番の手順を確定します。",
    finish: "本番の手順に迷いがなくなった。",
    point: "復習が終わったら新しい問題は解かず、早めに休みます。",
    tasks: [
      task("d28-note", "弱点ノートを見直す", 60, "昼のうちに確認する。"),
      task("d28-mock", "17:00 令和7年度秋期 最終模試", 150, "19:30まで本番と同じ時間帯で解く。", { exam: "r7a" }),
      task("d28-review", "模試の採点と復習", 90, "致命的なミスだけを確認し、深追いしない。", { exam: "r7a" }),
    ],
  },
  {
    date: "2026-11-29", dow: "日", phase: "仕上げ", title: "本番当日", exam: true,
    goal: "午前に軽く見直し、落ち着いて17:00の試験に臨みます。",
    finish: "試験を最後まで受け切る。",
    point: "開始直後に5問を選び、「選択」ボタンを押します。",
    tasks: [
      task("d29-note", "弱点ノートを見直す", 45, "午前中だけにして、新しい問題は解かない。"),
      task("d29-flow", "本番の手順を確認する", 15, "「本番」タブの時間配分と注意点を読む。"),
      task("d29-id", "本人確認書類を持って出発する", 0, "有効な本人確認書類がないと受験できない。"),
      task("d29-arrive", "16:30の受付開始を目安に到着する", 0, "集合時刻は試験開始15分前の16:45。"),
      task("d29-exam", "17:00 試験開始", 150, "19:30まで。全力を出し切りましょう。"),
    ],
  },
];

const PHASES = ["診断", "分野別演習", "模試と補強", "仕上げ"];
const MOCKS = [
  ["m1", "11/23 R6秋"],
  ["m2", "11/26 R7春"],
  ["m3", "11/28 R7秋"],
];

const params = new URLSearchParams(location.search);
const jstToday = new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Tokyo" }).format(new Date());
const today = /^\d{4}-\d{2}-\d{2}$/.test(params.get("date") || "") ? params.get("date") : jstToday;
const startDate = DAYS[0].date;
const examDate = DAYS.at(-1).date;

const STORAGE_KEY = "ap-study-navi:progress:v1";
const emptyProgress = () => ({ done: {}, tech: "", reserve: "", scores: {} });

let progress = emptyProgress();
let view = "today";
let selectedDate = today < startDate ? startDate : today > examDate ? examDate : today;

const app = document.querySelector("#app");
const statusEl = document.querySelector("#save-status");

const esc = (value) => String(value).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]);
const shortDate = (date) => `${Number(date.slice(5, 7))}/${Number(date.slice(8))}`;
const dayIndex = (date) => DAYS.findIndex((day) => day.date === date);
const daysBetween = (a, b) => Math.round((Date.parse(`${b}T00:00:00Z`) - Date.parse(`${a}T00:00:00Z`)) / 86400000);
const formatMin = (min) => {
  const hours = Math.floor(min / 60);
  const rest = min % 60;
  if (!hours) return `${rest}分`;
  return rest ? `${hours}時間${rest}分` : `${hours}時間`;
};
const techLabel = () => FIELD_NAMES[progress.tech] || "技術枠（11/20に決定）";
const reserveLabel = () => FIELD_NAMES[progress.reserve] || "予備分野（11/20に決定）";
const fill = (text) => text.replace("{tech}", techLabel()).replace("{reserve}", reserveLabel());
const isDone = (item) => progress.done[item.id] === true;
const dayStats = (day) => {
  const done = day.tasks.filter(isDone).length;
  const total = day.tasks.length;
  const totalMin = day.tasks.reduce((sum, item) => sum + item.min, 0);
  const leftMin = day.tasks.filter((item) => !isDone(item)).reduce((sum, item) => sum + item.min, 0);
  return { done, total, totalMin, leftMin, complete: done === total };
};

function setStatus(text, isError = false) {
  statusEl.textContent = text;
  statusEl.classList.toggle("error", isError);
}

function normalizeProgress(input) {
  const clean = emptyProgress();
  if (!input || typeof input !== "object") return clean;
  const knownTasks = new Set(DAYS.flatMap((day) => day.tasks.map((item) => item.id)));
  for (const [id, value] of Object.entries(input.done ?? {})) {
    if (knownTasks.has(id) && value === true) clean.done[id] = true;
  }
  if (["db", "nw", "pg"].includes(input.tech)) clean.tech = input.tech;
  if (["st", "db", "nw", "pg"].includes(input.reserve) && input.reserve !== clean.tech) clean.reserve = input.reserve;
  for (const [id] of MOCKS) {
    const score = input.scores?.[id];
    if (Number.isInteger(score) && score >= 0 && score <= 100) clean.scores[id] = score;
  }
  return clean;
}

function save() {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(progress));
    setStatus("保存済み");
  } catch {
    setStatus("保存できませんでした", true);
  }
}

function load() {
  try {
    progress = normalizeProgress(JSON.parse(localStorage.getItem(STORAGE_KEY)));
    setStatus("");
  } catch {
    progress = emptyProgress();
    setStatus("進捗を読み込めませんでした", true);
  }
}

function examLinks(key) {
  const [, path] = EXAMS[key];
  return `<a href="${IPA}${path}_ap_pm_qs.pdf" target="_blank" rel="noopener noreferrer">問題</a>
    <a href="${IPA}${path}_ap_pm_ans.pdf" target="_blank" rel="noopener noreferrer">解答例</a>
    <a href="${IPA}${path}_ap_pm_cmnt.pdf" target="_blank" rel="noopener noreferrer">採点講評</a>`;
}

function pickers() {
  const techOptions = ["db", "nw", "pg"].map((key) => `<option value="${key}" ${progress.tech === key ? "selected" : ""}>${FIELD_NAMES[key]}</option>`).join("");
  const reserveOptions = ["st", "db", "nw", "pg"].filter((key) => key !== progress.tech)
    .map((key) => `<option value="${key}" ${progress.reserve === key ? "selected" : ""}>${FIELD_NAMES[key]}</option>`).join("");
  return `<div class="inline-fields">
    <div class="field"><label for="pick-tech">技術枠</label><select id="pick-tech" data-pick="tech"><option value="">選択してください</option>${techOptions}</select></div>
    <div class="field"><label for="pick-reserve">予備</label><select id="pick-reserve" data-pick="reserve"><option value="">選択してください</option>${reserveOptions}</select></div>
  </div>`;
}

function taskItem(item, isNext) {
  const links = [
    item.exam ? examLinks(item.exam) : "",
    item.link ? `<a href="${item.link[0]}" target="_blank" rel="noopener noreferrer">${esc(item.link[1])}</a>` : "",
    item.link2 ? `<a href="${item.link2[0]}" target="_blank" rel="noopener noreferrer">${esc(item.link2[1])}</a>` : "",
  ].join("");
  return `<li class="task ${isDone(item) ? "done" : ""} ${isNext ? "next" : ""}">
    <input id="${item.id}" type="checkbox" data-task="${item.id}" ${isDone(item) ? "checked" : ""}>
    <div class="task-main">
      ${isNext ? '<span class="next-label">次にやること</span>' : ""}
      <label class="task-title" for="${item.id}">${esc(fill(item.title))}</label>
      <p class="task-detail">${esc(item.detail)}</p>
      ${links ? `<div class="links">${links}</div>` : ""}
      ${item.picker ? pickers() : ""}
    </div>
    <span class="task-time">${item.min ? formatMin(item.min) : ""}</span>
  </li>`;
}

function scorePanel(id, label) {
  const value = progress.scores[id];
  const has = Number.isInteger(value);
  const gap = has ? value - 60 : 0;
  const result = !has ? "" : gap >= 0
    ? `<p class="score-result good">合格基準を${gap}点上回っています</p>`
    : `<p class="score-result warn">合格基準まであと${-gap}点です</p>`;
  return `<div class="score-row"><label for="score-${id}">${esc(label)}（100点満点）</label>
    <input id="score-${id}" data-score="${id}" type="number" min="0" max="100" step="1" inputmode="numeric" value="${has ? value : ""}">${result}</div>`;
}

function dayStrip() {
  return `<div class="day-strip" role="group" aria-label="日付を選ぶ">${DAYS.map((day) => {
    const stats = dayStats(day);
    const state = stats.complete ? "完了" : day.exam ? "本番" : day.mock ? "模試" : stats.done ? `${stats.done}/${stats.total}` : "";
    return `<button type="button" class="day-chip ${day.date === today ? "is-today" : ""} ${day.mock || day.exam ? "mock" : ""}" data-date="${day.date}" aria-pressed="${day.date === selectedDate}">
      <small>${day.dow}</small><b>${shortDate(day.date)}</b><span class="state">${state}</span></button>`;
  }).join("")}</div>`;
}

function renderToday() {
  const day = DAYS[dayIndex(selectedDate)];
  const stats = dayStats(day);
  const nextTask = day.tasks.find((item) => !isDone(item));
  const relation = day.date === today ? "今日" : day.date > today ? `${daysBetween(today, day.date)}日後` : `${daysBetween(day.date, today)}日前`;
  let notice = "";
  if (today < startDate) notice = `<p class="notice">学習開始は${daysBetween(today, startDate) === 1 ? "明日" : `${daysBetween(today, startDate)}日後`}です。11/19の内容を確認しておきましょう。</p>`;
  else if (today > examDate) notice = '<p class="notice">試験お疲れさまでした。成績照会は2027年1月頃、合格発表は2027年2月頃の予定です。</p>';
  const pct = stats.total ? Math.round((stats.done / stats.total) * 100) : 0;
  app.innerHTML = `${notice}${dayStrip()}
    <div class="today-grid">
      <section aria-labelledby="day-title">
        <div class="hero">
          <p class="eyebrow">${shortDate(day.date)}（${day.dow}）・${relation}・${dayIndex(day.date) + 1}/${DAYS.length}日目<span class="tag ${day.mock || day.exam ? "mock" : ""}">${day.phase}</span></p>
          <h1 id="day-title">${esc(day.title)}</h1>
          <p>${esc(day.goal)}</p>
          <div class="progress-line">
            <div class="progress-track"><div class="progress-fill" data-width="${pct}"></div></div>
            <span class="progress-text">${stats.done}/${stats.total}完了${stats.leftMin ? `・残り${formatMin(stats.leftMin)}` : ""}</span>
          </div>
        </div>
        <ul class="task-list">${day.tasks.map((item) => taskItem(item, nextTask === item)).join("")}</ul>
        ${stats.complete ? '<p class="complete">この日の課題はすべて完了です。</p>' : ""}
      </section>
      <aside class="side" aria-label="この日の補足">
        <div class="panel"><h2>完了の目安</h2><p>${esc(day.finish)}</p></div>
        <div class="panel"><h2>今日のポイント</h2><p>${esc(day.point)}</p></div>
        ${day.mock ? `<div class="panel"><h2>模試の得点</h2>${scorePanel(day.mock, "得点を記録")}</div>` : ""}
      </aside>
    </div>`;
  app.querySelectorAll("[data-width]").forEach((el) => { el.style.width = `${el.dataset.width}%`; });
  app.querySelector('.day-chip[aria-pressed="true"]')?.scrollIntoView({ block: "nearest", inline: "center" });
}

function renderPlan() {
  const all = DAYS.flatMap((day) => day.tasks);
  const done = all.filter(isDone).length;
  const pct = Math.round((done / all.length) * 100);
  const totalMin = DAYS.slice(0, -1).reduce((sum, day) => sum + dayStats(day).totalMin, 0) + 60;
  app.innerHTML = `<section class="section">
      <div class="overall"><div><p class="eyebrow">11/19〜11/29・学習 約${Math.round(totalMin / 60)}時間</p><h1>5分野に絞って60点を取る</h1></div></div>
      <div class="progress-line"><div class="progress-track"><div class="progress-fill" data-width="${pct}"></div></div><span class="progress-text">${done}/${all.length}完了</span></div>
      <p class="task-detail">進捗はこの端末のブラウザに保存され、他の端末とは共有されません。</p>
      ${PHASES.map((phase) => `<div class="phase"><h3>${phase}</h3><ul class="plan-list">${DAYS.filter((day) => day.phase === phase).map((day) => {
        const stats = dayStats(day);
        return `<li><button type="button" class="plan-row ${day.date === today ? "is-today" : ""}" data-open="${day.date}">
          <span class="plan-date">${shortDate(day.date)}<br><small>${day.dow}</small></span>
          <span><span class="plan-title">${esc(day.title)}</span><span class="plan-meta">${day.exam ? "17:00〜19:30" : formatMin(stats.totalMin)}${day.mock ? "・模試" : ""}</span></span>
          <span class="plan-state ${stats.complete ? "done" : ""}">${stats.complete ? "完了" : `${stats.done}/${stats.total}`}</span></button></li>`;
      }).join("")}</ul></div>`).join("")}
    </section>
    <section class="section two-col">
      <div class="panel"><h2>解く5分野</h2>
        <table class="field-table"><tbody>
          <tr><th>必須</th><td>問1 情報セキュリティ</td></tr>
          <tr><th>主力</th><td>問9 プロジェクトマネジメント<br>問10 サービスマネジメント<br>問11 システム監査</td></tr>
          <tr><th>技術枠</th><td>${esc(techLabel())}</td></tr>
          <tr><th>予備</th><td>${esc(reserveLabel())}</td></tr>
        </tbody></table>${pickers()}
      </div>
      <div class="panel"><h2>模試の得点</h2><p>目標は各回60点以上です。</p>${MOCKS.map(([id, label]) => scorePanel(id, label)).join("")}</div>
    </section>`;
  app.querySelectorAll("[data-width]").forEach((el) => { el.style.width = `${el.dataset.width}%`; });
}

function renderExam() {
  const rows = [
    ["150→140分", "5問を選び、「選択」を押す"], ["140→115分", "1問目"], ["115→90分", "2問目"],
    ["90→65分", "3問目"], ["65→40分", "4問目"], ["40→15分", "5問目"], ["15→0分", "未解答の確認と見直し"],
  ];
  app.innerHTML = `<section class="section">
      <p class="eyebrow">11/29（日）17:00〜19:30</p><h1>本番の進め方</h1>
    </section>
    <section class="section">
      <div class="alert"><strong>開始直後に5問を選び、「選択」ボタンを押す</strong>押さずに終了すると選択が反映されず、採点されません。時間切れの場合はアラートも出ません。</div>
    </section>
    <section class="section"><h2>残り時間の目安</h2>
      <div class="time-bar" aria-hidden="true">
        <span class="buffer" data-flex="10">選択</span><span data-flex="25">1問目</span><span data-flex="25">2問目</span><span data-flex="25">3問目</span><span data-flex="25">4問目</span><span data-flex="25">5問目</span><span class="buffer" data-flex="15">確認</span>
      </div>
      <table class="time-table"><tbody>${rows.map(([time, label]) => `<tr><th>${time}</th><td>${label}</td></tr>`).join("")}</tbody></table>
      <p class="task-detail">1問25分を超えたら、途中でも次の問題へ進みます。</p>
    </section>
    <section class="section two-col">
      <div class="panel"><h2>CBTの注意点</h2><ul>
        <li>問題文はコピーできません。</li>
        <li>Ctrl+Cなどのキー操作は使えません。画面のボタンを使います。</li>
        <li>上限文字数を超えた入力は削除されます。</li>
        <li>入力方式は開始前に選び、入力欄で試します。</li>
        <li>メモ用紙とボールペンは会場で配布されたものを使います。</li>
      </ul></div>
      <div class="panel"><h2>記述のルール</h2><ul>
        <li>設問を先に読み、探す内容を決めます。</li>
        <li>本文の言葉を使って答えます。</li>
        <li>字数の8〜9割を埋めます。</li>
        <li>問われたことだけに答えます。</li>
      </ul></div>
    </section>`;
  app.querySelectorAll("[data-flex]").forEach((el) => { el.style.flex = el.dataset.flex; });
}

function render() {
  document.querySelectorAll("[data-view]").forEach((button) => {
    if (button.dataset.view === view) button.setAttribute("aria-current", "page");
    else button.removeAttribute("aria-current");
  });
  const left = daysBetween(today, examDate);
  document.querySelector("#countdown").textContent = left > 0 ? `本番まであと${left}日` : left === 0 ? "本日本番" : "試験終了";
  if (view === "plan") renderPlan();
  else if (view === "exam") renderExam();
  else renderToday();
}

document.addEventListener("click", (event) => {
  const viewButton = event.target.closest("[data-view]");
  const dateButton = event.target.closest("[data-date]");
  const openButton = event.target.closest("[data-open]");
  if (viewButton) {
    view = viewButton.dataset.view;
    render();
    scrollTo({ top: 0 });
  } else if (dateButton) {
    selectedDate = dateButton.dataset.date;
    render();
  } else if (openButton) {
    selectedDate = openButton.dataset.open;
    view = "today";
    render();
    scrollTo({ top: 0 });
  }
});

document.addEventListener("change", (event) => {
  const el = event.target;
  if (el.dataset.task) {
    if (el.checked) progress.done[el.dataset.task] = true;
    else delete progress.done[el.dataset.task];
  } else if (el.dataset.pick) {
    progress[el.dataset.pick] = el.value;
    if (progress.reserve && progress.reserve === progress.tech) progress.reserve = "";
  } else if (el.dataset.score) {
    const value = Number(el.value);
    if (el.value === "") delete progress.scores[el.dataset.score];
    else if (Number.isInteger(value) && value >= 0 && value <= 100) progress.scores[el.dataset.score] = value;
    else {
      setStatus("0〜100の整数で入力してください", true);
      return;
    }
  } else return;
  save();
  render();
});

window.addEventListener("storage", (event) => {
  if (event.key === STORAGE_KEY) {
    load();
    render();
  }
});

load();
render();
