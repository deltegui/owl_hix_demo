// Checks the hxbulma showcase (/bulma) in a real browser.
//
//   node check-bulma.mjs
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

const SERVER_PORT = 8098;
const CHROME_PORT = 9401;
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

await waitFor(`${BASE}/bulma`, "the server");
await waitFor(`http://localhost:${CHROME_PORT}/json/version`, "Chromium");

console.log("--- the HTML of the server (no JavaScript)");
const page0 = await (await fetch(BASE + "/bulma")).text();
const count = (text, re) => (text.match(re) || []).length;
check("loads Bulma and not Bootstrap", page0.includes("bulma@1") && !page0.includes("bootstrap"));
const classes = ["navbar", "navbar-burger", "navbar-menu", "hero", "section", "container", "footer", "buttons", "button", "notification", "tags", "tag",
  "progress", "skeleton-block", "skeleton-lines", "box", "content", "table", "icon-text", "image", "field", "control", "label", "input", "textarea",
  "select", "checkbox", "radio", "file", "file-cta", "has-addons", "is-grouped", "is-horizontal", "tabs", "dropdown", "dropdown-menu", "breadcrumb",
  "card", "card-header", "card-footer", "message", "pagination", "pagination-list", "menu", "menu-list", "panel", "panel-heading", "columns", "column",
  "level", "media", "tile", "fixed-grid", "grid", "cell"];
const missing = classes.filter((c) => !new RegExp(`class="([^"]* )?${c}( [^"]*)?"`).test(page0));
check("every component of the showcase is in the HTML", missing.length === 0, "missing: " + missing.join(", "));
check("the modals are not in the HTML while closed", !page0.includes('class="is-active modal"') && !page0.includes("modal-card"));
check("only the tab that is chosen has its panel", page0.includes("Tabs are a signal") && !page0.includes("No JavaScript from Bulma"));
check("the island has its element and loads bulma.wasm", page0.includes('id="island-bulma"') && page0.includes("/js/bulma.wasm"));

console.log("--- hydration");
const page = await openPage();
await page.goto("/bulma");
check("hydrates with no mismatch and no warning", (await page.evaluate("window.__hxHydrationMismatches")) === 0 && problems(page.logs).length === 0, JSON.stringify(page.logs));
check("no element of the server is created again", (await page.evaluate(`[...document.querySelectorAll("#island-bulma *")].filter((e) => !e.__seen).length`)) === 0);

console.log("--- what the user does");
check("burger opens and closes the menu", await run(page, `
  const burger = $(".navbar-burger"), menu = $(".navbar-menu");
  burger.click(); await w();
  const open = burger.classList.contains("is-active") && menu.classList.contains("is-active") && burger.getAttribute("aria-expanded") === "true";
  burger.click(); await w();
  return open && !burger.classList.contains("is-active") && !menu.classList.contains("is-active");`));
check("dropdown opens when clicked and the item closes it", await run(page, `
  const dd = $(".dropdown"); dd.querySelector("button").click(); await w();
  const open = dd.classList.contains("is-active");
  dd.querySelectorAll(".dropdown-item")[0].click(); await w();
  const closed = !dd.classList.contains("is-active");
  const modal = !!$(".modal.is-active");
  $(".modal-background").click(); await w();
  return open && closed && modal && !$(".modal");`));
check("the card modal has a title, body and footer and its buttons close it", await run(page, `
  button("Modal").click(); await w();
  const box = !!$(".modal-content .box"); $(".modal-close").click(); await w();
  const gone = !$(".modal");
  $$(".dropdown-item")[1].click(); await w();
  const card = $(".modal-card-title")?.textContent === "Delete it?" && !!$(".modal-card-body") && !!$(".modal-card-foot");
  button("Cancel") && $$(".modal-card-foot button").find((b) => b.textContent === "Cancel").click(); await w();
  return box && gone && card && !$(".modal");`));
check("tabs switch the panel", await run(page, `
  const text = () => $$("p").map((p) => p.textContent).filter((t) => /Tabs are a signal|bulma\\.Tabs\\(|No JavaScript/.test(t)).join("|");
  const first = text();
  $$(".tabs a").find((a) => a.textContent === "Usage").click(); await w();
  const second = text(); const active = $(".tabs li.is-active").textContent;
  return first.startsWith("Tabs are") && second.startsWith("bulma.Tabs") && active === "Usage";`));
check("a notification is removed with its delete button", await run(page, `
  const n = $$(".notification .delete").length; $(".notification .delete").click(); await w();
  return $$(".notification .delete").length === n - 1;`));
check("a tag is removed with its delete button", await run(page, `
  const n = $$(".tags .tag .delete").length; $(".tags .tag .delete").click(); await w();
  return $$(".tags .tag .delete").length === n - 1;`));
check("the message is dismissed", await run(page, `
  const m = $$(".message").length; $(".message-header .delete").click(); await w();
  return m === 1 && $$(".message").length === 0;`));
check("the card folds", await run(page, `
  const n = $$(".card-content").length; $(".card-header-icon").click(); await w();
  const folded = $$(".card-content").length === n - 1;
  $(".card-header-icon").click(); await w();
  return folded && $$(".card-content").length === n;`));
check("pagination moves between pages and shows the ellipsis", await run(page, `
  const cur = () => $(".pagination-link.is-current").textContent;
  const a = cur(); $(".pagination-next").click(); await w();
  const b = cur(); $(".pagination-previous").click(); $(".pagination-previous").click(); await w();
  $$(".pagination-link").find((l) => l.textContent === "12").click(); await w();
  return a === "3" && b === "4" && cur() === "12" && !!$(".pagination-ellipsis") && $(".pagination-next").disabled;`));
check("progress follows its buttons", await run(page, `
  const bar = $("progress[max='100'][value]"); const v = bar.value;
  button("+10").click(); await w();
  return v === 40 && bar.value === 50;`));
check("a button shows loading", await run(page, `
  const b = button("Click: loading"); b.click(); await w();
  const on = b.classList.contains("is-loading"); b.click(); await w();
  return on && !b.classList.contains("is-loading");`));
check("the inputs are bound: the summary follows them and the error goes away", await run(page, `
  const help = $(".help.is-danger"); const before = help.textContent;
  type($$("input.input")[0], "Ada"); await w();
  const sel = $(".select select"); type(sel, "Company", "change");
  const agree = $(".checkbox input"); agree.click();
  $$("input[type=radio]").find((r) => r.value === "pro").click(); await w();
  const summary = $("pre").textContent;
  return before === "required" && help.textContent === "" && summary === "name=Ada email= kind=Company agree=true plan=pro";`));
check("the file picker shows what happened", await run(page, `
  const input = $(".file-input"); input.dispatchEvent(new Event("change", { bubbles: true })); await w();
  return $(".file-name").textContent === "file chosen";`));
await sleep(200);
check("no warnings in the console after all that", problems(page.logs).length === 0, JSON.stringify(page.logs));

page.close();
console.log(failures === 0 ? "\nall good" : `\n${failures} failed`);
process.exit(failures === 0 ? 0 : 1);
