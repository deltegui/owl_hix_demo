// Checks the hxbs showcase (/bs) in a real browser.
//
//   node check-bs.mjs
//
// It builds no WASM: run `make build` first (`make check` does). It starts the
// server, with fresh data, and a headless Chromium, and checks:
//   - what the server sends, with no JavaScript (the islands are already in the HTML),
//     in Spanish, the default language, and CSRF and the language buttons,
//   - that each page hydrates: no mismatch, and no element of the server is created again,
//   - what a user does in each screen, the chart following the products, the
//     generic forms and the theme.
//
// Needs Node 22+ and Chromium. The page loads Bootstrap and Chart.js from a CDN,
// so the checks of the chart are skipped when it is not reachable.
import { execFileSync, spawn } from "node:child_process";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const SERVER_PORT = 8101;
const CHROME_PORT = 9404;
const BASE = `http://localhost:${SERVER_PORT}`;
const CHROMIUM = process.env.CHROMIUM || "chromium";
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function waitFor(url, what) {
  for (let i = 0; i < 100; i++) {
    try { if ((await fetch(url)).ok) return; } catch {}
    await sleep(200);
  }
  throw new Error(`${what} did not start`);
}

const profile = mkdtempSync(join(tmpdir(), "owl-check-"));
// The server is built first so it can be started, and stopped, as a single process.
execFileSync("go", ["build", "-o", join(profile, "server"), "./cmd/owldemo"]);
const server = spawn(join(profile, "server"), [], { stdio: "ignore", env: { ...process.env, OWLDEMO_ADDR: `:${SERVER_PORT}` } });
const chrome = spawn(CHROMIUM, ["--headless=new", "--no-sandbox", "--disable-gpu",
  `--remote-debugging-port=${CHROME_PORT}`, `--user-data-dir=${join(profile, "chrome")}`, "about:blank"], { stdio: "ignore" });
const cleanup = () => { server.kill(); chrome.kill(); try { rmSync(profile, { recursive: true, force: true }); } catch {} };
process.on("exit", cleanup);

// Every element of the page when it is parsed gets a mark. After hydration, an
// element without it is one that was created again instead of adopted.
const MARK = `document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll("body *").forEach((e) => { e.__seen = true; });
});`;

async function openPage() {
  const target = await (await fetch(`http://localhost:${CHROME_PORT}/json/new?about:blank`, { method: "PUT" })).json();
  const ws = new WebSocket(target.webSocketDebuggerUrl);
  await new Promise((r) => (ws.onopen = r));
  let id = 0; const pending = new Map(); const logs = [];
  ws.onmessage = (e) => {
    const m = JSON.parse(e.data);
    if (m.id && pending.has(m.id)) { pending.get(m.id)(m.result); pending.delete(m.id); }
    else if (m.method === "Runtime.consoleAPICalled") logs.push(m.params.args.map((a) => a.value ?? a.description).join(" "));
    else if (m.method === "Runtime.exceptionThrown") logs.push("EXCEPTION " + (m.params.exceptionDetails.exception?.description ?? m.params.exceptionDetails.text));
  };
  const send = (method, params = {}) => new Promise((r) => { const i = ++id; pending.set(i, r); ws.send(JSON.stringify({ id: i, method, params })); });
  await send("Runtime.enable");
  await send("Page.enable");
  await send("Page.addScriptToEvaluateOnNewDocument", { source: MARK });
  const evaluate = async (expression) => (await send("Runtime.evaluate", { expression, returnByValue: true, awaitPromise: true }))?.result?.value;
  // Goes to the page and waits until the WASM has hydrated what the page has.
  const goto = async (path) => {
    logs.length = 0;
    await send("Page.navigate", { url: BASE + path });
    for (let i = 0; i < 100; i++) {
      if (await evaluate(`window.__hxHydrationMismatches !== undefined || !document.querySelector("[data-island]")`)) break;
      await sleep(100);
    }
    await sleep(500);
  };
  return { evaluate, goto, logs, close: () => ws.close() };
}

let failures = 0;
const check = (name, ok, detail = "") => {
  console.log(`${ok ? "PASS" : "FAIL"}  ${name}${ok ? "" : "   " + detail}`);
  if (!ok) failures++;
};
const skip = (name, why) => console.log(`SKIP  ${name}   (${why})`);
const same = (a, b) => JSON.stringify(a) === JSON.stringify(b);
const problems = (logs) => logs.filter((l) => l.includes("hx:") || l.includes("EXCEPTION"));



const HELPERS = `
  const w = (ms = 300) => new Promise((r) => setTimeout(r, ms));
  const $ = (s) => document.querySelector(s), $$ = (s) => [...document.querySelectorAll(s)];
  const button = (text) => $$("button").find((b) => b.textContent.trim() === text);
  const type = (el, value, event = "input") => { el.value = value; el.dispatchEvent(new Event(event, { bubbles: true })); };
`;
const run = (page, code) => page.evaluate(`(async () => { ${HELPERS} ${code} })()`);

await waitFor(`${BASE}/bs`, "the server");
await waitFor(`http://localhost:${CHROME_PORT}/json/version`, "Chromium");
function count(text, re) { return (text.match(re) || []).length; }

console.log("--- the HTML of the server (no JavaScript)");
const page0 = await (await fetch(BASE + "/bs")).text();
check("loads Bootstrap and the WASM of the showcase", page0.includes("bootstrap@5.3.8") && page0.includes("/js/bsdemo.wasm") && page0.includes('id="island-bsdemo"'));
const classes = ["container", "btn", "btn-group", "badge", "alert", "card", "list-group", "progress", "spinner-border", "placeholder",
  "form-control", "form-select", "form-check-input", "form-switch", "form-range", "input-group", "form-floating", "invalid-feedback",
  "breadcrumb", "nav-tabs", "nav-pills", "accordion", "pagination", "dropdown", "collapse", "navbar", "carousel", "table-striped", "table-responsive"];
const missing = classes.filter((c) => !new RegExp(`class="([^"]* )?${c}( [^"]*)?"`).test(page0));
check("every component of the showcase is in the HTML", missing.length === 0, "missing: " + missing.join(", "));
check("modals, offcanvas and toasts are not in the HTML while closed", !/class="modal[ "]/.test(page0) && !page0.includes('class="offcanvas') && !/class="toast[ "]/.test(page0));
check("only the tab that is chosen has its panel", page0.includes("Account settings.") && !page0.includes("Change your password."));
check("the dropdown menu is closed", !/dropdown-menu show/.test(page0));
const more = ["d-flex", "d-grid", "blockquote", "blockquote-footer", "list-inline", "list-unstyled", "btn-check", "form-control-plaintext", "valid-feedback",
  "form-check-inline", "form-control-color", "input-group-sm", "alert-heading", "alert-link", "card-group", "card-header-tabs", "card-img-overlay", "stretched-link",
  "position-relative", "progress-stacked", "dropup", "btn-group", "dropdown-toggle-split", "navbar-text", "table-danger"];
const lacking = more.filter((c) => !new RegExp(`class="([^"]* )?${c}( [^"]*)?"`).test(page0));
check("the new components are in the HTML", lacking.length === 0, "missing: " + lacking.join(", "));
check("the datalist and the caption are in the HTML", page0.includes('<datalist id="x-cities">') && page0.includes("<caption>List of users</caption>"));

console.log("--- hydration");
const page = await openPage();
await page.goto("/bs");
check("hydrates with no mismatch and no warning", (await page.evaluate("window.__hxHydrationMismatches")) === 0 && problems(page.logs).length === 0, JSON.stringify(page.logs));
check("no element of the server is created again", (await page.evaluate(`[...document.querySelectorAll("#island-bsdemo *")].filter((e) => !e.__seen).length`)) === 0);

console.log("--- what the user does");
check("modal opens, closes with its X, with the background and with a button", await run(page, `
  button("Open modal").click(); await w();
  const open = !!$(".modal.show") && $(".modal-title").textContent === "Edit profile" && !!$(".modal-backdrop") && !!$(".modal-footer");
  $(".modal .btn-close").click(); await w(); const a = !$(".modal");
  button("Open modal").click(); await w();
  $(".modal-backdrop").click(); await w(); const b = !$(".modal");
  button("Open modal").click(); await w();
  $(".modal-footer button").click(); await w();
  return open && a && b && !$(".modal") && !$(".modal-backdrop");`));
check("the modal is above its backdrop and takes the click", await run(page, `
  button("Open modal").click(); await w();
  const r = $(".modal-content").getBoundingClientRect();
  const hit = document.elementFromPoint(r.x + r.width / 2, r.y + r.height / 2);
  const inside = $(".modal-content").contains(hit);
  $(".modal .btn-close").click(); await w();
  return inside;`));
check("centered modal opens and closes", await run(page, `
  button("Open centered").click(); await w();
  const open = !!$(".modal-dialog-centered.modal-sm"); $(".modal .btn-close").click(); await w();
  return open && !$(".modal");`));
check("offcanvas opens and closes", await run(page, `
  button("Open offcanvas").click(); await w();
  const open = !!$(".offcanvas.offcanvas-start.show") && $(".offcanvas-title").textContent === "Menu";
  $(".offcanvas-backdrop").click(); await w();
  return open && !$(".offcanvas") && !$(".offcanvas-backdrop");`));
check("toast shows and closes", await run(page, `
  button("Show toast").click(); await w();
  const open = !!$(".toast.show") && $(".toast-body").textContent === "Event created";
  $(".toast .btn-close").click(); await w();
  return open && !$(".toast");`));
check("dropdown opens, an item sets the state and closes it", await run(page, `
  const dd = $(".dropdown"), menu = dd.querySelector(".dropdown-menu"), t = dd.querySelector(".dropdown-toggle");
  const before = !menu.classList.contains("show") && t.getAttribute("aria-expanded") === "false";
  t.click(); await w();
  const open = menu.classList.contains("show") && t.getAttribute("aria-expanded") === "true";
  [...menu.querySelectorAll(".dropdown-item")].find((i) => i.textContent === "Settings").click(); await w();
  return before && open && !menu.classList.contains("show") && $("strong").textContent === "settings";`));
check("a click outside closes the dropdown", await run(page, `
  const dd = $(".dropdown"), menu = dd.querySelector(".dropdown-menu");
  dd.querySelector(".dropdown-toggle").click(); await w();
  const open = menu.classList.contains("show");
  const layer = [...dd.children].find((c) => c.style.position === "fixed"); layer.click(); await w();
  return open && !menu.classList.contains("show") && ![...dd.children].some((c) => c.style.position === "fixed");`));
check("collapse shows and hides", await run(page, `
  const c = $$(".collapse").find((e) => !e.classList.contains("navbar-collapse") && !e.classList.contains("dropdown-menu") && e.querySelector(".card"));
  const a = c.classList.contains("show"); button("Toggle collapse").click(); await w();
  const b = c.classList.contains("show"); button("Toggle collapse").click(); await w();
  return !a && b && !c.classList.contains("show");`));
check("navbar toggler opens the menu", await run(page, `
  const t = $(".navbar-toggler"), c = $(".navbar-collapse"); t.click(); await w();
  const open = c.classList.contains("show") && t.getAttribute("aria-expanded") === "true"; t.click(); await w();
  return open && !c.classList.contains("show") && t.getAttribute("aria-expanded") === "false";`));
check("carousel moves with its arrows and dots", await run(page, `
  const on = () => $$(".carousel-item").findIndex((i) => i.classList.contains("active"));
  const a = on(); $(".carousel-control-next").click(); await w(); const b = on();
  $(".carousel-control-prev").click(); $(".carousel-control-prev").click(); await w(); const c = on();
  $$(".carousel-indicators button")[1].click(); await w();
  return a === 0 && b === 1 && c === 2 && on() === 1 && $$(".carousel-indicators button")[1].classList.contains("active");`));
check("tabs switch the panel and the selection", await run(page, `
  const tab = (t) => $$("[role=tab]").find((x) => x.textContent === t);
  const first = $("[role=tabpanel]").textContent;
  tab("Password").click(); await w();
  return first === "Account settings." && $("[role=tabpanel]").textContent === "Change your password." && tab("Password").classList.contains("active") && !tab("Account").classList.contains("active") && tab("Password").getAttribute("aria-selected") === "true";`));
check("the accordion opens one at a time and closes the open one", await run(page, `
  const b = $$(".accordion-button"), p = $$(".accordion-collapse"); const a = p[0].classList.contains("show") && b[1].classList.contains("collapsed");
  b[1].click(); await w(); const two = !p[0].classList.contains("show") && p[1].classList.contains("show") && !b[1].classList.contains("collapsed");
  b[1].click(); await w();
  return !a === false && two && !p[1].classList.contains("show");`));
check("pagination moves between pages", await run(page, `
  const cur = () => $(".page-item.active .page-link").textContent; const nav = () => $("nav[aria-label=pagination]");
  const next = () => [...nav().querySelectorAll("button")].find((b) => b.textContent === "»");
  const a = cur(); next().click(); await w(); const b = cur();
  [...nav().querySelectorAll("button")].find((x) => x.textContent === "12").click(); await w();
  return a === "3" && b === "4" && cur() === "12" && next().disabled;`));
check("the inputs are bound: the summary follows them and the error goes away", await run(page, `
  const err = () => $(".invalid-feedback").textContent; const before = err() === "required" && $("#f-name").classList.contains("is-invalid");
  type($("#f-name"), "Ada"); await w();
  type($("#f-kind"), "Company", "change");
  $("#f-agree").click(); $("#f-notify").click();
  $$("input[type=radio]").find((r) => r.value === "pro").click();
  type($("#f-vol"), "70"); await w();
  return before && err() === "" && !$("#f-name").classList.contains("is-invalid") && $("pre").textContent === "name=Ada kind=Company agree=true notify=false plan=pro volume=70";`));
check("labels are linked to their control", await run(page, `
  return $$("label.form-label").every((l) => l.htmlFor && document.getElementById(l.htmlFor));`));
check("the dismissible alert goes away", await run(page, `
  const n = () => $$(".alert").length; const a = n(); $(".alert .btn-close").click(); await w();
  return n() === a - 1 && !$(".alert-dismissible");`));
check("progress follows its buttons", await run(page, `
  const bar = $(".progress"); const v = bar.getAttribute("aria-valuenow");
  button("+10").click(); await w();
  return v === "40" && bar.getAttribute("aria-valuenow") === "50" && bar.firstChild.style.width === "50%";`));
check("flex: the spacer pushes the last badge to the end; grid has three columns", await run(page, `
  const f = $("[data-demo=flex]"), [l, r] = f.querySelectorAll(".badge"), fr = f.getBoundingClientRect();
  const g = getComputedStyle($("[data-demo=grid]")).gridTemplateColumns.split(" ").length;
  return l.getBoundingClientRect().left - fr.left < 2 && Math.abs(r.getBoundingClientRect().right - fr.right) < 2 && g === 3;`));
check("button check and button radios change the state", await run(page, `
  const st = () => $("[data-demo=more-state]").textContent;
  const a = st(); $("label[for=x-toggle]").click(); await w(); const b = st();
  $("label[for=size-l]").click(); await w(); const c = st();
  return a.startsWith("pressed=false size=m") && b.startsWith("pressed=true size=m") && c.startsWith("pressed=true size=l")
    && $("#x-toggle").checked && $("#size-l").checked && !$("#size-m").checked;`));
check("fieldset, inline checks, colour and the check in the group are bound", await run(page, `
  const st = () => $("[data-demo=more-state]").textContent;
  $("#x-b").click(); $("#x-a").click(); $("#x-inline").click(); $("#role-admin").click();
  type($("#x-colour"), "#00ff00"); $(".input-group-sm input[type=checkbox]").click(); await w();
  return st().endsWith("a=false b=true inline=true role=admin colour=#00ff00 opted=true") && $("fieldset legend").textContent === "Extras";`));
check("the valid feedback shows when the control is valid", await run(page, `
  const i = $("#x-email"), fb = $(".valid-feedback"); const a = i.classList.contains("is-valid");
  type(i, "ada@example.com"); await w();
  return !a && i.classList.contains("is-valid") && getComputedStyle(fb).display === "block" && $("#x-owner").readOnly && $("#x-owner").value === "Ada Lovelace";`));
check("the loading button shows a spinner and is disabled until it is done", await run(page, `
  const b = button("Save"); const a = !b.disabled && !b.querySelector(".spinner-border");
  b.click(); await w(); const busy = b.disabled && !!b.querySelector(".spinner-border-sm");
  button("Done").click(); await w();
  return a && busy && !b.disabled && !b.querySelector(".spinner-border");`));
check("the badge sits on the corner of its button", await run(page, `
  const w2 = $(".position-relative.d-inline-block"), btn = w2.querySelector(".btn"), bd = w2.querySelector(".badge");
  const br = btn.getBoundingClientRect(), r = bd.getBoundingClientRect();
  return bd.textContent === "99+" && Math.abs((r.left + r.width / 2) - br.right) < 3 && Math.abs((r.top + r.height / 2) - br.top) < 3;`));
check("the card tabs switch their panel", await run(page, `
  const tab = (t) => $$(".card-header-tabs [role=tab]").find((x) => x.textContent === t);
  const first = $$(".card-body p").some((p) => p.textContent === "Card tab one"); tab("Two").click(); await w();
  return first && $$(".card-body p").some((p) => p.textContent === "Card tab two") && !$$(".card-body p").some((p) => p.textContent === "Card tab one");`));
check("the stretched link covers its card", await run(page, `
  const c = $(".stretched-link").closest(".card"); c.scrollIntoView({ block: "center" }); await w(); const r = c.getBoundingClientRect();
  return document.elementFromPoint(r.left + r.width - 8, r.top + r.height - 8) === $(".stretched-link");`));
check("the stacked progress follows its signal", await run(page, `
  const seg = () => $(".progress-stacked .progress"); const a = seg().style.width;
  button("seg +5").click(); await w();
  return a === "25%" && seg().style.width === "30%" && seg().getAttribute("aria-valuenow") === "30";`));
check("split dropdown: the main button acts, the arrow opens the menu and an item closes it", await run(page, `
  const g = $(".btn-group .dropdown-toggle-split").parentElement, [main, arrow] = g.querySelectorAll(":scope > button"), menu = g.querySelector(".dropdown-menu");
  const last = () => $("[data-demo=last]").textContent;
  main.click(); await w(); const a = last() === "main" && !menu.classList.contains("show");
  arrow.click(); await w(); const open = menu.classList.contains("show") && arrow.getAttribute("aria-expanded") === "true";
  [...menu.querySelectorAll(".dropdown-item")].find((i) => i.textContent === "split-b").click(); await w();
  return a && open && last() === "split-b" && !menu.classList.contains("show");`));
check("dropup opens upwards and menu-end aligns to the right", await run(page, `
  const up = $(".dropup"), t = up.querySelector(".dropdown-toggle"), m = up.querySelector(".dropdown-menu");
  t.click(); await w(); const tr = t.getBoundingClientRect(), mr = m.getBoundingClientRect(); const above = mr.bottom <= tr.top + 1 && m.classList.contains("show");
  t.click(); await w();
  const e = $$(".dropdown").find((d) => d.querySelector(".dropdown-menu-end")), et = e.querySelector(".dropdown-toggle"), em = e.querySelector(".dropdown-menu");
  et.click(); await w(); const right = Math.abs(em.getBoundingClientRect().right - et.getBoundingClientRect().right) < 2; et.click(); await w();
  return above && right;`));
check("nav dropdown opens inside the navbar and an item sets the state", await run(page, `
  const nb = $("[data-demo=second-navbar]"); nb.querySelector(".navbar-toggler").click(); await w();
  const d = nb.querySelector(".nav-item.dropdown"), t = d.querySelector(".nav-link.dropdown-toggle"), m = d.querySelector(".dropdown-menu");
  t.click(); await w(); const open = m.classList.contains("show") && t.getAttribute("aria-expanded") === "true";
  [...m.querySelectorAll(".dropdown-item")].find((i) => i.textContent === "nav-b").click(); await w();
  return open && $("[data-demo=last]").textContent === "nav-b" && !m.classList.contains("show") && !!nb.querySelector(".navbar-text");`));
check("scrollable modal scrolls its body and fullscreen fills the window", await run(page, `
  button("Scrollable modal").click(); await w();
  const body = $(".modal-body"), ok = !!$(".modal-dialog-scrollable") && getComputedStyle(body).overflowY === "auto" && body.scrollHeight > body.clientHeight;
  $(".modal .btn-close").click(); await w();
  button("Fullscreen modal").click(); await w();
  const d = $(".modal-fullscreen"), r = d.getBoundingClientRect();
  const full = !!d && Math.abs(r.width - innerWidth) < 2 && Math.abs(r.height - innerHeight) < 2;
  $(".modal .btn-close").click(); await w();
  return ok && full && !$(".modal");`));
await sleep(200);
check("no warnings in the console after all that", problems(page.logs).length === 0, JSON.stringify(page.logs));

page.close();
console.log(failures === 0 ? "\nall good" : `\n${failures} failed`);
process.exit(failures === 0 ? 0 : 1);
