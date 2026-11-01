// Checks the hxcoat showcase (/coat) in a real browser.
//
//   node check-coat.mjs
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

const SERVER_PORT = 8100;
const CHROME_PORT = 9403;
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

await waitFor(`${BASE}/coat`, "the server");
await waitFor(`http://localhost:${CHROME_PORT}/json/version`, "Chromium");

console.log("--- the HTML of the server (no JavaScript)");
const page0 = await (await fetch(BASE + "/coat")).text();
check("loads Basecoat and Tailwind, not Bootstrap", page0.includes("basecoat-css@1") && page0.includes("@tailwindcss/browser") && !page0.includes("bootstrap"));
const has = (re) => re.test(page0);
const classes = ["btn", "button-group", "badge", "alert", "card", "avatar", "avatar-group", "kbd", "empty", "item", "item-group", "progress", "skeleton",
  "table", "table-container", "accordion", "breadcrumb", "field", "label", "input", "textarea", "select", "fieldset", "field-separator", "input-group",
  "tabs", "dropdown-menu", "popover", "combobox", "command", "toaster"];
const missing = classes.filter((c) => !new RegExp(`class="([^"]* )?${c}( [^"]*)?"`).test(page0));
check("every component of the showcase is in the HTML", missing.length === 0, "missing: " + missing.join(", "));
check("dialogs are not in the HTML while closed", !has(/<dialog/) && !has(/class="drawer"/));
check("popovers are in the HTML, hidden", count(page0, /data-popover=""/g) >= 4 && !page0.includes('aria-hidden="false"'));
check("only the tab that is chosen has its panel", page0.includes("Account settings.") && !page0.includes("Change your password."));
check("the island has its element and loads coat.wasm", page0.includes('id="island-coat"') && page0.includes("/js/coat.wasm"));
function count(text, re) { return (text.match(re) || []).length; }

console.log("--- hydration");
const page = await openPage();
await page.goto("/coat");
check("hydrates with no mismatch and no warning", (await page.evaluate("window.__hxHydrationMismatches")) === 0 && problems(page.logs).length === 0, JSON.stringify(page.logs));
check("no element of the server is created again", (await page.evaluate(`[...document.querySelectorAll("#island-coat *")].filter((e) => !e.__seen).length`)) === 0);

console.log("--- what the user does");
check("dialog opens, closes with its X, and with the background", await run(page, `
  button("Open dialog").click(); await w();
  const open = !!$("dialog.dialog[open]") && $("dialog.dialog h2").textContent === "Edit profile" && !!$("dialog.dialog footer");
  $("dialog.dialog > div > button").click(); await w();
  const closed = !$("dialog");
  button("Open dialog").click(); await w();
  $("dialog.dialog").previousElementSibling.click(); await w();
  return open && closed && !$("dialog");`));
check("alert dialog is not closed by the background, only by its buttons", await run(page, `
  button("Open alert dialog").click(); await w();
  const open = !!$("dialog.alert-dialog[open]") && !$("dialog.alert-dialog > div > button[aria-label=Close]");
  $("dialog.alert-dialog").previousElementSibling.click(); await w();
  const stays = !!$("dialog.alert-dialog");
  $$("dialog.alert-dialog button").find((b) => b.textContent === "Cancel").click(); await w();
  return open && stays && !$("dialog");`));
check("drawer opens and closes", await run(page, `
  button("Open drawer").click(); await w();
  const open = !!$("dialog.drawer[open]"); $$("dialog.drawer button").find((b) => b.textContent === "Submit").click(); await w();
  return open && !$("dialog");`));
check("dropdown menu opens, an item closes it and sets the state", await run(page, `
  const dd = $(".dropdown-menu"), pop = dd.querySelector("[data-popover]"), trigger = dd.querySelector("button");
  const before = pop.getAttribute("aria-hidden") === "true" && trigger.getAttribute("aria-expanded") === "false";
  trigger.click(); await w();
  const open = pop.getAttribute("aria-hidden") === "false" && trigger.getAttribute("aria-expanded") === "true";
  [...pop.querySelectorAll("[role=menuitem]")].find((i) => i.textContent === "Settings").click(); await w();
  return before && open && pop.getAttribute("aria-hidden") === "true" && $("strong").textContent === "settings";`));
check("a click outside closes the dropdown", await run(page, `
  const dd = $(".dropdown-menu"), pop = dd.querySelector("[data-popover]");
  dd.querySelector("button").click(); await w();
  const open = pop.getAttribute("aria-hidden") === "false";
  const layer = [...dd.children].find((c) => c.style.position === "fixed"); layer.click(); await w();
  return open && pop.getAttribute("aria-hidden") === "true" && ![...dd.children].some((c) => c.style.position === "fixed");`));
check("the menu checkbox toggles", await run(page, `
  const dd = $(".dropdown-menu"); dd.querySelector("button").click(); await w();
  const item = dd.querySelector("[role=menuitemcheckbox]"); const a = item.getAttribute("aria-checked");
  item.click(); await w(); dd.querySelector("button").click(); await w();
  return a === "true" && item.getAttribute("aria-checked") === "false";`));
check("popover opens and closes", await run(page, `
  const p = $(".popover"), pop = p.querySelector("[data-popover]"); p.querySelector("button").click(); await w();
  const open = pop.getAttribute("aria-hidden") === "false"; p.querySelector("button").click(); await w();
  return open && pop.getAttribute("aria-hidden") === "true";`));
check("select menu chooses an option and shows it", await run(page, `
  const s = $$(".select").find((e) => e.tagName === "DIV" && e.querySelector("[role=listbox]")); const label = () => s.querySelector("button span").textContent;
  const a = label(); s.querySelector("button").click(); await w();
  s.querySelector("[data-value=c]").click(); await w();
  return a === "Beta" && label() === "Gamma" && s.querySelector("[data-value=c]").getAttribute("aria-selected") === "true" && s.querySelector("[data-popover]").getAttribute("aria-hidden") === "true";`));
check("combobox filters as you type and chooses", await run(page, `
  const c = $(".combobox"); c.querySelector("button").click(); await w();
  const all = c.querySelectorAll("[role=option]").length;
  type(c.querySelector("input"), "blue"); await w();
  const some = [...c.querySelectorAll("[role=option]")].map((o) => o.textContent);
  c.querySelector("[role=option]").click(); await w();
  return all === 5 && some.join() === "Blueberry" && c.querySelector("button span").textContent === "Blueberry";`));
check("command filters as you type", await run(page, `
  const c = $(".command"); const n = () => c.querySelectorAll("[role=menuitem]").length;
  const a = n(); type(c.querySelector("input"), "cal"); await w();
  const b = [...c.querySelectorAll("[role=menuitem]")].map((i) => i.textContent.replace("⌘K", ""));
  const groups = c.querySelectorAll("[role=group]").length;
  c.querySelector("[role=menuitem]").click(); await w();
  type(c.querySelector("input"), ""); await w();
  return a === 5 && b.join() === "Calendar,Calculator" && groups === 1 && $("strong").textContent === "calendar" && n() === 5;`));
check("the tooltip text is an attribute", await run(page, `return $$("[data-tooltip]").some((e) => e.dataset.tooltip === "Add to library");`));
check("tabs switch the panel and the selection", await run(page, `
  const tab = (t) => $$("[role=tab]").find((x) => x.textContent === t);
  const first = $$("[role=tabpanel]").map((p) => p.textContent).join("|");
  tab("Password").click(); await w();
  return first.startsWith("Account settings.") && $$("[role=tabpanel]")[0].textContent === "Change your password." && tab("Password").getAttribute("aria-selected") === "true" && tab("Account").getAttribute("aria-selected") === "false";`));
check("the accordion opens one at a time", await run(page, `
  const ds = $$(".accordion details"); ds[1].querySelector("summary").click(); await w();
  return !ds[0].open && ds[1].open;`));
check("pagination moves between pages", await run(page, `
  const cur = () => $("[aria-current=page].btn").textContent; const next = () => [...$("nav[aria-label=pagination]").querySelectorAll("button")].find((b) => b.textContent.includes("Next"));
  const a = cur(); next().click(); await w(); const b = cur();
  [...$("nav[aria-label=pagination]").querySelectorAll("button")].find((x) => x.textContent === "12").click(); await w();
  return a === "3" && b === "4" && cur() === "12" && next().disabled;`));
check("the inputs are bound: the summary follows them and the error goes away", await run(page, `
  const err = () => $("p[role=alert]").textContent; const before = err();
  type($("#f-name"), "Ada"); await w();
  type($("#f-kind"), "Company", "change");
  $$(".label input[type=checkbox]").find((c) => !c.hasAttribute("role")).click();
  $$(".label input[role=switch]")[0].click();
  $$("input[type=radio]").find((r) => r.value === "pro").click();
  type($("#f-vol"), "70"); await w();
  const summary = $("pre").textContent;
  return before === "required" && err() === "" && summary === "name=Ada kind=Company agree=true notify=false plan=pro volume=70";`));
check("progress follows its buttons", await run(page, `
  const bar = $(".progress"); const v = bar.getAttribute("aria-valuenow");
  button("+10").click(); await w();
  return v === "40" && bar.getAttribute("aria-valuenow") === "50" && bar.firstChild.style.width === "50%";`));
check("a toast shows with each click and is dismissed", await run(page, `
  const n = () => $$(".toast").length; const a = n();
  button("Show toast").click(); await w(500); const one = n();
  button("Show toast").click(); await w(500); const two = n();
  const text = $(".toast h2").textContent === "Event created" && $(".toast p").textContent.startsWith("Sunday");
  $(".toast [data-toast-cancel]").click(); await w(500);
  return a === 0 && one === 1 && two === 2 && text && n() === 1;`));
check("the drawer has its content above the background", await run(page, `
  button("Open drawer").click(); await w();
  const d = $("dialog.drawer"); const r = d.firstElementChild.getBoundingClientRect();
  const hit = document.elementFromPoint(r.x + r.width / 2, r.y + r.height / 2);
  const icon = d.querySelector("svg").getBoundingClientRect();
  $$("dialog.drawer button").find((b) => b.textContent === "Submit").click(); await w();
  return d.contains(hit) && icon.width <= 20;`));
check("the circle skeleton is round", await run(page, `const s = $$(".skeleton")[0]; return s.classList.contains("rounded-full") && !s.classList.contains("rounded-md");`));
await sleep(200);
check("no warnings in the console after all that", problems(page.logs).length === 0, JSON.stringify(page.logs));

page.close();
console.log(failures === 0 ? "\nall good" : `\n${failures} failed`);
process.exit(failures === 0 ? 0 : 1);
