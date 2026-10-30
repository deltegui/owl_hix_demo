// Checks the OWL demo in a real browser.
//
//   node check.mjs
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

const SERVER_PORT = 8097;
const CHROME_PORT = 9400;
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

// What the page does in the browser, written once so each script reads like a user.
const HELPERS = `
  const w = (ms = 400) => new Promise((r) => setTimeout(r, ms));
  const $ = (s) => document.querySelector(s), $$ = (s) => [...document.querySelectorAll(s)];
  const button = (text, root = "") => $$(root + " button").find((b) => b.textContent.trim() === text);
  const type = (el, value, event = "input") => { el.value = value; el.dispatchEvent(new Event(event, { bubbles: true })); };
  const rows = () => $$("#island-products tbody tr, #island-contacts tbody tr").map((r) => [...r.children].slice(1, 3).map((c) => c.textContent.trim()).join("|"));
  const status = () => $(".text-muted.small")?.textContent;
  const errors = () => $$(".invalid-feedback").map((e) => e.textContent).filter(Boolean);
`;

await waitFor(`${BASE}/about`, "the server");
await waitFor(`http://localhost:${CHROME_PORT}/json/version`, "Chromium");

// --- What the server sends, with no JavaScript ---------------------------------

console.log("--- the HTML of the server");
const html = async (path) => (await fetch(BASE + path)).text();
const count = (text, re) => (text.match(re) || []).length;
// The binary a page loads is the one its script fetches; a page may name the others in its text.
const bundleOf = (text) => (text.match(/fetch\("\/js\/(\w+)\.wasm"\)/) || [])[1];

const products = await html("/products");
check("products: the table is in the HTML (8 rows) with the total, in Spanish", products.includes("<td><span>Atlas 1</span>") && count(products, /<tr>/g) === 9 && products.includes("30 encontrados") && products.includes("<h1>Productos</h1>"));
check("products: the state of each island is in the page", products.includes('id="island-products-state"') && products.includes('id="island-products-stats-state"'));
check("products: loads full.wasm and Chart.js", bundleOf(products) === "full" && products.includes("chart.js"));
check("the menu comes from the menu feature, in Spanish, with the current page active", products.includes('class="active nav-link text-white" href="/products">Productos<') && products.includes(">Contactos<"));
const contacts = await html("/contacts");
check("contacts: 8 rows and the total", count(contacts, /<tr>/g) === 9 && contacts.includes("20 encontrados"));
check("contacts: loads full.wasm, no Chart.js", bundleOf(contacts) === "full" && !contacts.includes("chart.js"));
const entriesPage = await html("/entries");
check("entries: the list is in the HTML (8 rows, header and total) in Spanish", count(entriesPage, /<tr>/g) === 9 && entriesPage.includes("<h1>Entradas del libro</h1>") && entriesPage.includes("40 encontrados") && entriesPage.includes("Total:"));
check("entries: its state is in the page and it loads entries.wasm", entriesPage.includes('id="island-entries-state"') && bundleOf(entriesPage) === "entries");
const formPage = await html("/forms/product");
check("a form page already has the form (the labels)", ["Nombre", "Categoría", "Precio"].every((l) => formPage.includes(">" + l + "</label>")) && formPage.includes("Activo"));
check("a form page loads forms.wasm", bundleOf(formPage) === "forms");
check("an unknown form is a 404", (await fetch(BASE + "/forms/nothing")).status === 404);
check("a request to nothing is a 404 and not an empty 200", (await fetch(BASE + "/api/nothing")).status === 404);
const about = await html("/about");
check("about: plain text, loads only lite.wasm", bundleOf(about) === "lite");
check("about: has the theme selector rendered", about.includes('id="island-theme"') && about.includes("<select"));
check("the page has the CSRF token and the language buttons", /<meta content="[^"]+" name="csrf-token"\/>/.test(about) && about.includes('action="/language/en"'));

// --- OWL: CSRF and the language form -----------------------------------------

console.log("--- csrf and language");
const tokenOf = (text) => (text.match(/<meta content="([^"]+)" name="csrf-token"/) || [])[1];
const token = tokenOf(about);
const send = (path, init) => fetch(BASE + path, { redirect: "manual", ...init });
const json = { "Content-Type": "application/json", Accept: "application/json" };
const newProduct = JSON.stringify({ name: "Zeppelin", category: "Games", price: 12.5, active: true });
check("a POST with no CSRF token is refused", (await send("/api/products", { method: "POST", headers: json, body: newProduct })).status === 403);
check("a POST with a made up token is refused", (await send("/api/products", { method: "POST", headers: { ...json, "X-Csrf-Token": "nope" }, body: newProduct })).status === 403);
const created = await send("/api/products", { method: "POST", headers: { ...json, "X-Csrf-Token": token }, body: newProduct });
const createdId = (await created.json()).id;
check("with the token it is accepted", created.status === 200 && createdId > 30, String(createdId));
check("a DELETE needs it too, and with it the product goes", (await send(`/api/products/${createdId}`, { method: "DELETE" })).status === 403 &&
  (await send(`/api/products/${createdId}`, { method: "DELETE", headers: { "X-Csrf-Token": token } })).status === 204);
check("a GET needs no token", (await send("/api/products?name=Atlas")).status === 200);
const bad = await send("/api/products", { method: "POST", headers: { ...json, "X-Csrf-Token": token }, body: JSON.stringify({ name: "Ab", category: "", price: 0 }) });
const badBody = await bad.json();
check("validation: 422, by field, translated (Spanish by default), and it says it is JSON", bad.status === 422 && bad.headers.get("content-type").startsWith("application/json") && badBody.errors.name[0] === "El campo Nombre debe tener al menos 3 caracteres" && badBody.errors.category[0] === "El campo Categoría es obligatorio" && badBody.errors.price[0] === "El campo Precio debe valer al menos 0.01", JSON.stringify(badBody));
const form = (next) => new URLSearchParams({ "X-Csrf-Token": token, Next: next });
const toEn = await send("/language/en", { method: "POST", body: form("/contacts?x=1") });
check("the language form redirects to the page it came from", toEn.status === 303 && toEn.headers.get("location") === "/contacts?x=1", toEn.headers.get("location"));
const toEvil = await send("/language/en", { method: "POST", body: form("//evil.example/") });
check("and never to another site", toEvil.headers.get("location") === "/", toEvil.headers.get("location"));
check("the language form needs the token too", (await send("/language/en", { method: "POST", body: new URLSearchParams({ Next: "/" }) })).status === 403);
const cookie = (toEn.headers.getSetCookie?.() ?? []).map((c) => c.split(";")[0]).join("; ");
const english = await (await fetch(BASE + "/products", { headers: { cookie } })).text();
check("with the cookie, the page is in English", english.includes("<h1>Products</h1>") && english.includes("30 found") && english.includes('lang="en"'));

// --- Hydration: nothing is created again -----------------------------------

console.log("--- hydration");
const page = await openPage();
const ADOPTED = `(() => { const all = [...document.querySelectorAll("body *")]; return { total: all.length, recreated: all.filter((e) => !e.__seen).length }; })()`;
for (const path of ["/products", "/contacts", "/entries", "/forms/product", "/forms/contact", "/about"]) {
  await page.goto(path);
  const r = await page.evaluate(ADOPTED);
  check(`${path}: no mismatch, no error in the console`, problems(page.logs).length === 0, problems(page.logs).join(" | "));
  check(`${path}: the ${r.total} elements of the server are the ones in use`, r.recreated === 0, `${r.recreated} were created again`);
}

// --- Products: the screen (in English: the language button of the page) ----------------------------------------------------

console.log("--- products");
await page.goto("/products");
await page.evaluate(`document.querySelector('form[action="/language/en"] button').click()`);
await sleep(1500);
await page.goto("/products");
let r = await page.evaluate(`(async () => { ${HELPERS}
  const out = {};
  out.first = rows().length; out.status = status();
  out.pages = $$(".pagination .page-link").map((b) => b.textContent).join(",");
  $$(".pagination .page-link").find((b) => b.textContent === "2").click(); await w(600);
  out.page2 = rows()[0];
  type($$("main form input[type=text]")[0], "atlas"); $("main form button[type=submit]").click(); await w(600);
  out.filtered = rows(); out.filteredStatus = status();
  button("Clear").click(); await w(600); out.cleared = rows().length;
  button("New product").click(); await w(300);
  out.modal = !!$(".modal.show");
  button("Save", ".modal-footer").click(); await w(700);
  out.errors = errors(); out.stillOpen = !!$(".modal.show");
  type($$(".modal-body input[type=text]")[0], "Zeppelin"); type($(".modal-body select"), "Games", "change"); type($(".modal-body input[type=number]"), "12.5");
  button("Save", ".modal-footer").click(); await w(800);
  out.closed = !$(".modal.show"); out.notice = $(".alert")?.textContent; out.total = status();
  button("Edit", "tbody tr:first-child").click(); await w(300);
  out.edit = [$$(".modal-body input[type=text]")[0].value, $(".modal-body select").value, $(".modal-body input[type=number]").value];
  $$(".modal-header button")[0].click(); await w(200);
  const badge = $("tbody tr:first-child .badge").textContent;
  $$("tbody tr:first-child button").find((b) => /Disable|Enable/.test(b.textContent)).click(); await w(700);
  out.toggle = badge + " -> " + $("tbody tr:first-child .badge").textContent;
  button("Delete", "tbody tr:first-child").click(); await w(300);
  out.confirm = $(".modal-body")?.textContent;
  button("Delete", ".modal-footer").click(); await w(800);
  out.after = rows()[0]; out.deleted = $(".alert")?.textContent;
  return out; })()`);
check("the first page has 8 rows, the total and the pages", r.first === 8 && r.status === "30 found" && r.pages === "«,1,2,3,»", JSON.stringify([r.first, r.status, r.pages]));
check("page 2 loads from the API", r.page2 === "Bellows 1|Books", r.page2);
check("the filter works", same(r.filtered, ["Atlas 1|Books", "Atlas 2|Music", "Atlas 3|Books"]) && r.filteredStatus === "3 found", JSON.stringify(r.filtered));
check("clearing the filter goes back", r.cleared === 8);
check("saving an empty form shows the errors and keeps the modal", r.modal && r.stillOpen && same(r.errors, ["The Name field is required", "The Category field is required", "The Price field must be at least 0.01"]), JSON.stringify(r.errors));
check("creating a product closes the modal, says it and counts it", r.closed && r.notice.includes("Product saved.") && r.total === "31 found", JSON.stringify([r.closed, r.notice, r.total]));
check("editing opens with the values of the row", same(r.edit, ["Atlas 1", "Books", "5.99"]), JSON.stringify(r.edit));
check("enable and disable work", r.toggle === "Active -> Disabled", r.toggle);
check("delete asks first, then deletes", r.confirm === "Delete Atlas 1? This cannot be undone." && r.after === "Compass 1|Games" && r.deleted.includes("Product deleted."), JSON.stringify([r.confirm, r.after, r.deleted]));
check("no errors in the console", problems(page.logs).length === 0, problems(page.logs).join(" | "));

// --- The chart follows the products (two islands, one WASM) ----------------

console.log("--- chart");
await page.goto("/products");
if (!(await page.evaluate(`typeof Chart !== "undefined"`))) {
  skip("the chart follows the products", "Chart.js could not be loaded from the CDN");
} else {
  r = await page.evaluate(`(async () => { ${HELPERS}
    const chart = () => Chart.getChart($("#island-products-stats canvas"));
    const data = () => chart().data.datasets[0].data.join(",");
    const first = chart(); const out = { before: data() };
    const toggle = () => $$("#island-products tbody tr:first-child button").find((b) => /Disable|Enable/.test(b.textContent)).click();
    toggle(); await w(700); out.disabled = data();
    toggle(); await w(700); out.enabled = data();
    button("New product").click(); await w(300);
    type($$(".modal-body input[type=text]")[0], "Blimp"); type($(".modal-body select"), "Games", "change"); type($(".modal-body input[type=number]"), "100");
    button("Save", ".modal-footer").click(); await w(900); out.created = data();
    button("New product").click(); await w(300);
    button("Save", ".modal-footer").click(); await w(700); out.invalid = data();
    $$(".modal-header button")[0].click(); await w(200);
    out.sameChart = chart() === first; out.noRefreshButton = !$$("#island-products-stats button").length;
    return out; })()`);
  const before = r.before.split(",").map(Number), created = r.created.split(",").map(Number);
  check("the chart was drawn when the page hydrated", before.length === 4 && before.every((n) => n > 0), r.before);
  check("disabling a product changes the chart and enabling it brings it back", r.disabled !== r.before && r.enabled === r.before, JSON.stringify([r.before, r.disabled, r.enabled]));
  check("a new Games product of 100 adds exactly 100 to Games", Math.abs(created[1] - before[1] - 100) < 0.001 && created[0] === before[0], r.created);
  check("a rejected form changes nothing", r.invalid === r.created, r.invalid);
  check("it is the same chart all the time, with no refresh button", r.sameChart && r.noRefreshButton);
}

// --- Contacts: the server decides ----------------------------------------------

console.log("--- contacts");
await page.goto("/contacts");
r = await page.evaluate(`(async () => { ${HELPERS}
  button("New contact").click(); await w(300);
  const [name, email] = $$(".modal-body input[type=text]"); type(name, "Ana Lopez"); type(email, "abcdefg");
  button("Save", ".modal-footer").click(); await w(800);
  return { errors: errors(), stillOpen: !!$(".modal.show"), status: status() }; })()`);
check("a rule only the server checks (the email has no @) shows its error", r.stillOpen && r.errors.includes("The Email field must contain @"), JSON.stringify(r.errors));
check("the contacts screen is the same code with another configuration", r.status === "20 found", r.status);

// --- Generic forms: no request for the form, it comes with the page --------

console.log("--- forms");
await page.goto("/forms/product");
r = await page.evaluate(`(async () => { ${HELPERS}
  const out = { requests: performance.getEntriesByType("resource").map((e) => e.name.replace(location.origin, "")).filter((n) => n.includes("/api/")) };
  out.labels = $$("#island-form label").map((l) => l.textContent.trim());
  const submit = () => $("#island-form button[type=submit]").click();
  submit(); await w(300); out.empty = errors();
  const inputs = $$("#island-form input[type=text], #island-form input[type=number]");
  type(inputs[0], "Ab"); type(inputs[1], "0"); type($("#island-form select"), "Music", "change");
  submit(); await w(300); out.short = errors();
  type(inputs[0], "Kazoo"); type(inputs[1], "3.5"); submit(); await w(900);
  out.notice = $("#island-form .alert-success")?.textContent;
  out.created = await (await fetch("/api/products?name=Kazoo")).json();
  return out; })()`);
check("the form needs no request for its description", r.requests.length === 0, r.requests.join(", "));
check("the fields come from the struct", same(r.labels, ["Name", "Category", "Price", "Active"]), JSON.stringify(r.labels));
check("the rules of the struct are checked before sending", same(r.empty, ["This field is required", "This field is required", "Must be at least 0.01"]) && r.short[0] === "Must be at least 3", JSON.stringify([r.empty, r.short]));
check("a valid form is saved and the server has it", r.notice?.startsWith("Saved with id") && r.created.items.length === 1 && r.created.items[0].category === "Music" && r.created.items[0].price === 3.5, JSON.stringify([r.notice, r.created.items]));

await page.goto("/forms/contact");
r = await page.evaluate(`(async () => { ${HELPERS}
  const out = { labels: $$("#island-form label").map((l) => l.textContent.trim()) };
  const [name, email] = $$("#island-form input[type=text]"); type(name, "Ana Lopez"); type(email, "abcdefg");
  $("#island-form button[type=submit]").click(); await w(800); out.server = errors();
  type(email, "ana@example.com"); $("#island-form button[type=submit]").click(); await w(800);
  out.notice = $("#island-form .alert-success")?.textContent;
  return out; })()`);
check("another struct, the same island and the same WASM", same(r.labels, ["Name", "Email", "Phone", "Active"]), JSON.stringify(r.labels));
check("the server's 422 shows on its field, and then it saves", r.server.includes("The Email field must contain @") && r.notice?.startsWith("Saved with id"), JSON.stringify([r.server, r.notice]));


// --- Ledger entries: a maintenance with two layers --------------------------

console.log("--- entries");
await page.goto("/entries");
const ENTRY_HELPERS = HELPERS + `
  const list = () => $("[data-layer=list]");
  const listRows = () => $$("[data-layer=list] tbody tr");
  const rowIds = () => listRows().map((r) => r.children[0].textContent.trim());
  const listHidden = () => list().classList.contains("d-none");
  const editor = () => $("[data-layer=editor]");
  const sel = (field) => editor().querySelector("[data-field=" + field + "]");
  const phrase = () => $("[data-role=phrase]");
  const phraseSelects = () => [...phrase().querySelectorAll("select")];
  const accountsIn = (side) => [...editor().querySelector("select[data-side=" + side + "]").options].map((o) => o.value).filter(Boolean);
  const lineRows = () => $$("[data-layer=editor] tbody tr");
  const sumText = (role) => $("[data-role=" + role + "]")?.textContent;
  const click = (el) => el.click();
  const api = async (q) => (await fetch("/api/entries?" + q)).json();
`;
r = await page.evaluate(`(async () => { ${ENTRY_HELPERS}
  const out = {};
  out.rows = listRows().length; out.status = status();
  out.total = $("[data-layer=list] tfoot td:nth-child(2)").textContent;
  out.sum = (await api("page=1")).sum.toFixed(2) + " €";
  out.headers = $$("[data-layer=list] thead th").map((h) => h.textContent.trim());
  out.first = (await api("page=1")).items.map((i) => String(i.id));
  out.firstShown = rowIds();
  // Sorting: clicking a header sorts by it and shows the arrow; again, it turns around.
  $("th[data-sort=samount]").click(); await w(700);
  out.asc = rowIds(); out.ascHeader = $("th[data-sort=samount]").textContent.trim();
  out.ascApi = (await api("orderBy=samount&order=asc&page=1")).items.map((i) => String(i.id));
  $("th[data-sort=samount]").click(); await w(700);
  out.desc = rowIds(); out.descHeader = $("th[data-sort=samount]").textContent.trim();
  out.descApi = (await api("orderBy=samount&order=desc&page=1")).items.map((i) => String(i.id));
  $("th[data-sort=carriedOutAt]").click(); await w(500); $("th[data-sort=carriedOutAt]").click(); await w(700);
  // Pagination keeps the order.
  $$(".pagination .page-link").find((b) => b.textContent === "2").click(); await w(700);
  out.page2 = rowIds(); out.page2Api = (await api("orderBy=carriedOutAt&order=desc&page=2")).items.map((i) => String(i.id));
  return out; })()`);
check("entries: first page, header, total and the same rows the API gives", r.rows === 8 && r.status === "40 found" && r.total === r.sum && same(r.first, r.firstShown) && r.headers[0] === "#", JSON.stringify([r.rows, r.status, r.total, r.sum, r.first, r.firstShown]));
check("entries: a header sorts, with its arrow, and again it turns around", same(r.asc, r.ascApi) && r.ascHeader.endsWith("▲") && same(r.desc, r.descApi) && r.descHeader.endsWith("▼") && !same(r.asc, r.desc), JSON.stringify([r.asc, r.ascApi, r.ascHeader, r.desc, r.descApi]));
check("entries: page 2 keeps the order", same(r.page2, r.page2Api) && r.page2.length === 8, JSON.stringify([r.page2, r.page2Api]));

r = await page.evaluate(`(async () => { ${ENTRY_HELPERS}
  const out = {};
  type($$("[data-layer=list] form input[type=text]")[0], "grocer"); $("[data-layer=list] form button[type=submit]").click(); await w(700);
  out.concept = rowIds(); out.conceptStatus = status();
  out.conceptApi = (await api("concept=grocer&orderBy=carriedOutAt&order=desc&page=1")).items.map((i) => String(i.id));
  button("Clear").click(); await w(600);
  type($$("[data-layer=list] form select")[0], "3", "change"); type($$("[data-layer=list] form input[type=date]")[0], "2026-02-01"); type($$("[data-layer=list] form input[type=date]")[1], "2026-03-01");
  $("[data-layer=list] form button[type=submit]").click(); await w(700);
  out.range = status(); out.rangeTotal = $("[data-layer=list] tfoot td:nth-child(2)").textContent;
  out.rangeApi = await api("category=3&from=2026-02-01&to=2026-03-01&page=1");
  button("Clear").click(); await w(600);
  out.back = status();
  return out; })()`);
check("entries: the filter by concept matches the API", same(r.concept, r.conceptApi) && r.concept.length > 0 && r.conceptStatus === `${r.conceptApi.length} found`, JSON.stringify([r.concept, r.conceptApi, r.conceptStatus]));
check("entries: category and dates filter, and the total follows the filter", r.range === `${r.rangeApi.total} found` && r.rangeApi.total > 0 && r.rangeTotal === r.rangeApi.sum.toFixed(2) + " €" && r.back === "40 found", JSON.stringify([r.range, r.rangeApi.total, r.rangeTotal, r.rangeApi.sum, r.back]));

// View, over the list.
r = await page.evaluate(`(async () => { ${ENTRY_HELPERS}
  const out = {};
  type($$("[data-layer=list] form input[type=text]")[0], "Salary after taxes"); $("[data-layer=list] form button[type=submit]").click(); await w(700);
  button("View", "[data-layer=list] tbody tr:first-child").click(); await w(700);
  out.hidden = listHidden(); out.title = $("[data-layer=viewer] h1")?.textContent;
  out.lines = $$("[data-layer=viewer] tbody tr").length;
  out.names = $$("[data-layer=viewer] tbody tr td:first-child").map((c) => c.textContent);
  out.footer = $$("[data-layer=viewer] tfoot td").slice(1, 3).map((c) => c.textContent);
  out.created = $("[data-layer=viewer] .text-muted.small")?.textContent;
  button("Back").click(); await w(300);
  out.visible = !listHidden(); out.filterKept = $$("[data-layer=list] form input[type=text]")[0].value;
  button("Clear").click(); await w(600);
  return out; })()`);
check("entries: view covers the list and shows the lines with the full name of each account", r.hidden && r.title.startsWith("Entry #") && r.lines === 3 && r.names[0].startsWith("Asset · Bank") && r.names[1].startsWith("Expense · Taxes") && r.footer[0] === r.footer[1] && /^\d+\.\d\d €$/.test(r.footer[0]), JSON.stringify(r));
check("entries: back shows the list as it was (the filter is kept)", r.visible && r.filterKept === "Salary after taxes" && /Created by demo/.test(r.created), JSON.stringify([r.visible, r.filterKept, r.created]));

// Edit in basic mode: the sentence, and the accounts that follow the operation.
r = await page.evaluate(`(async () => { ${ENTRY_HELPERS}
  const out = {};
  type($$("[data-layer=list] form input[type=text]")[0], "Dinner out"); $("[data-layer=list] form button[type=submit]").click(); await w(700);
  out.id = rowIds()[0];
  button("Edit", "[data-layer=list] tbody tr:first-child").click(); await w(800);
  out.title = $("[data-layer=editor] h1")?.textContent;
  out.operation = sel("operation").value; out.amount = sel("amount").value;
  out.credit = accountsIn("credit"); out.debit = accountsIn("debit");
  out.creditValue = editor().querySelector("select[data-side=credit]").value; out.debitValue = editor().querySelector("select[data-side=debit]").value;
  out.mode = $("[data-role=mode]").textContent;
  // Operation: income, the accounts become income and asset ones.
  type(sel("operation"), "1", "change"); await w(400);
  out.incomeCredit = accountsIn("credit"); out.incomeDebit = accountsIn("debit");
  out.incomeCreditValue = editor().querySelector("select[data-side=credit]").value; out.incomeDebitValue = editor().querySelector("select[data-side=debit]").value;
  type(sel("operation"), "2", "change"); await w(400);
  type(sel("amount"), "99.5"); button("Save").click(); await w(900);
  out.notice = $(".alert-success")?.textContent; out.backToList = !listHidden();
  out.saved = await (await fetch("/api/entries/" + out.id)).json();
  button("Clear").click(); await w(600);
  return out; })()`);
check("entries: editing a basic entry shows its sentence", r.title === "Edit entry #" + r.id && r.operation === "2" && r.amount !== "" && r.mode === "Advanced edit (lines)", JSON.stringify([r.title, r.operation, r.amount, r.mode]));
check("entries: the accounts of each side are the ones of the group the operation asks for", same(r.credit, ["100", "101", "102", "103", "104"]) && same(r.debit, ["500", "501", "502"]) && r.creditValue === "100" && r.debitValue === "500", JSON.stringify([r.credit, r.debit, r.creditValue, r.debitValue]));
check("entries: changing the operation changes the accounts to ones that fit", same(r.incomeCredit, ["400", "401", "402", "403", "404"]) && same(r.incomeDebit, ["100", "101", "102", "103", "104"]) && r.incomeCreditValue === "400" && r.incomeDebitValue === "100", JSON.stringify([r.incomeCredit, r.incomeDebit, r.incomeCreditValue, r.incomeDebitValue]));
check("entries: saving goes back to the list, says it, and the server has it", r.backToList && r.notice.includes("saved") && r.saved.amount === 99.5 && r.saved.operationId === 2, JSON.stringify([r.backToList, r.notice, r.saved.amount, r.saved.operationId]));

// New entry: the server's rules show on their fields.
r = await page.evaluate(`(async () => { ${ENTRY_HELPERS}
  const out = {};
  button("New entry").click(); await w(800);
  out.title = $("[data-layer=editor] h1")?.textContent; out.operation = sel("operation").value;
  button("Save").click(); await w(800);
  out.errors = errors().concat($$("[data-layer=editor] .invalid-feedback div").map((d) => d.textContent).filter(Boolean));
  out.stillOpen = !!editor();
  type(sel("category"), "3", "change");
  type(sel("amount"), "250"); $("[data-tag='2']").click(); $("[data-tag='5']").click();
  type($("[data-layer=editor] textarea"), "A new salary");
  button("Save").click(); await w(900);
  out.notice = $(".alert-success")?.textContent; out.status = status();
  const created = (await api("concept=A+new+salary&page=1")).items[0];
  out.created = created; out.full = await (await fetch("/api/entries/" + created.id)).json();
  return out; })()`);
check("entries: a new entry opens as an income and an empty sentence says what is missing", r.title === "New entry" && r.operation === "1" && r.errors.includes("The Category field is required") && r.errors.includes("Field mandatory") && r.stillOpen, JSON.stringify([r.title, r.operation, r.errors, r.stillOpen]));
check("entries: after filling it, it is created with its category and its tags", r.notice.includes("saved") && r.status === "41 found" && r.created.category === "Salary" && r.created.tags === "monthly, deductible" && r.full.amount === 250 && r.full.debitAccount === 100 && r.full.creditAccount === 400, JSON.stringify([r.notice, r.status, r.created, r.full]));

// Advanced mode: lines, the sums, the currency and the balance.
r = await page.evaluate(`(async () => { ${ENTRY_HELPERS}
  const out = {};
  type($$("[data-layer=list] form input[type=text]")[0], "A new salary"); $("[data-layer=list] form button[type=submit]").click(); await w(700);
  const id = rowIds()[0];
  button("Edit", "[data-layer=list] tbody tr:first-child").click(); await w(800);
  $("[data-role=mode]").click(); await w(900);
  out.lines = lineRows().length; out.mode = $("[data-role=mode]").textContent;
  out.sums = [sumText("sum-debit"), sumText("sum-credit")];
  out.first = [lineRows()[0].querySelector("[data-field=account]").value, lineRows()[0].querySelector("[data-side=debit] input").value, lineRows()[1].querySelector("[data-side=credit] input").value];
  out.hiddenOther = lineRows()[0].querySelector("[data-side=credit]").classList.contains("d-none");
  $("[data-role=add-line]").click(); await w(300);
  out.added = lineRows().length;
  const third = () => lineRows()[2];
  type(third().querySelector("[data-side=debit] input"), "50"); await w(200);
  out.unbalanced = [sumText("sum-debit"), sumText("sum-credit")];
  out.red = $("[data-role=sum-debit]").classList.contains("text-danger");
  out.hidesCredit = third().querySelector("[data-side=credit]").classList.contains("d-none");
  button("Save").click(); await w(500);
  out.refused = $(".alert-danger")?.textContent; out.stillEditor = !!editor();
  // Balanced: another 50 on the credit side, in dollars.
  $("[data-role=add-line]").click(); await w(300);
  const fourth = () => lineRows()[3];
  type(fourth().querySelector("select[data-field=account]"), "401", "change");
  type(fourth().querySelector("[data-side=credit] input"), "50");
  type(fourth().querySelector("select[data-field=currency]"), "USD", "change"); await w(900);
  out.dollars = fourth().querySelector("[data-role=samount]").textContent;
  out.balancedNow = [sumText("sum-debit"), sumText("sum-credit")];
  out.green = $("[data-role=sum-credit]").classList.contains("text-success");
  type(third().querySelector("[data-side=debit] input"), String(parseFloat(out.dollars))); await w(200);
  out.after = [sumText("sum-debit"), sumText("sum-credit")];
  out.greenAgain = $("[data-role=sum-debit]").classList.contains("text-success");
  out.id = id;
  return out; })()`);
const dollars = parseFloat(r.dollars);
check("entries: the advanced mode shows the two lines of the entry, one side each", r.lines === 2 && r.mode === "Basic edit (sentence)" && r.sums[0] === r.sums[1] && r.first[0] === "100" && r.first[1] === "250" && r.first[2] === "250" && r.hiddenOther, JSON.stringify([r.lines, r.mode, r.sums, r.first, r.hiddenOther]));
check("entries: typing in one side hides the other, and the sums turn red when they differ", r.added === 3 && r.hidesCredit && r.unbalanced[0] === "300.00 €" && r.unbalanced[1] === "250.00 €" && r.red, JSON.stringify([r.added, r.hidesCredit, r.unbalanced, r.red]));
check("entries: it does not send an unbalanced entry", r.refused === "Debit and credit must add up to the same." && r.stillEditor, JSON.stringify([r.refused, r.stillEditor]));
check("entries: a line in dollars is priced with the rate the server gave, and the sums follow", dollars > 40 && dollars < 60 && dollars !== 50 && r.after[0] === r.after[1] && r.greenAgain, JSON.stringify([r.dollars, r.balancedNow, r.after, r.greenAgain]));

r = await page.evaluate(`(async () => { ${ENTRY_HELPERS}
  const out = {};
  button("Save").click(); await w(1000);
  out.notice = $(".alert-success")?.textContent; out.backToList = !listHidden();
  const entry = await (await fetch("/api/entries/" + ${JSON.stringify(r.id)})).json();
  out.saved = { mode: entry.mode, can: entry.canEditBasic, lines: entry.lines?.length, currencies: entry.lines?.map((l) => l.currency).join(",") };
  button("Edit", "[data-layer=list] tbody tr:first-child").click(); await w(800);
  out.mode = $("[data-role=mode]").textContent; out.disabled = $("[data-role=mode]").disabled;
  out.why = $("[data-layer=editor] p.text-muted.small")?.textContent;
  out.lines = lineRows().length;
  button("Back").click(); await w(300);
  return out; })()`);
check("entries: the balanced entry is saved with its four lines, in two currencies", r.backToList && r.notice.includes("saved") && r.saved.mode === "advanced" && r.saved.can === false && r.saved.lines === 4 && r.saved.currencies === "EUR,EUR,EUR,USD", JSON.stringify([r.backToList, r.notice, r.saved]));
check("entries: it opens in advanced mode, and the server says why it cannot be a sentence", r.mode === "Basic edit (sentence)" && r.disabled === true && r.why.includes("cannot be edited as a sentence") && r.lines === 4, JSON.stringify([r.mode, r.disabled, r.why, r.lines]));

// A rule of the server: the amounts do not balance (the browser check is skipped on purpose).
const token2 = tokenOf(await (await fetch(BASE + "/entries")).text());
const unbalanced = await fetch(BASE + "/api/entries", { method: "POST", headers: { "Content-Type": "application/json", "X-Csrf-Token": token2, Accept: "application/json", cookie: "" }, body: JSON.stringify({
  mode: "advanced", categoryId: 1, date: "2026-03-01", hour: 10, minute: 0, exchangeDate: "2026-03-01",
  lines: [{ account: 100, debit: 10, currency: "EUR" }, { account: 400, credit: 11, currency: "EUR" }] }) });
const unbalancedBody = await unbalanced.json();
check("entries: the server refuses an entry that does not balance, with a translated message", unbalanced.status === 400 && unbalancedBody.error.includes("El debe y el haber deben sumar lo mismo"), JSON.stringify([unbalanced.status, unbalancedBody]));
const oneSide = await fetch(BASE + "/api/entries", { method: "POST", headers: { "Content-Type": "application/json", "X-Csrf-Token": token2, Accept: "application/json" }, body: JSON.stringify({
  mode: "advanced", categoryId: 1, date: "2026-03-01", hour: 10, minute: 0,
  lines: [{ account: 100, debit: 10, credit: 10, currency: "EUR" }, { account: 400, currency: "EUR" }] }) });
const oneSideBody = await oneSide.json();
check("entries: a line must have either debit or credit, and the error is on its field", oneSide.status === 422 && oneSideBody.errors["lines[0].debit"] && oneSideBody.errors["lines[0].credit"] && oneSideBody.errors["lines[1].debit"], JSON.stringify(oneSideBody));

// Delete, with a confirmation.
r = await page.evaluate(`(async () => { ${ENTRY_HELPERS}
  const out = {};
  button("Clear").click(); await w(700);
  const before = status();
  const id = rowIds()[0];
  button("Delete", "[data-layer=list] tbody tr:first-child").click(); await w(300);
  out.confirm = $(".modal-body")?.textContent;
  button("Delete", ".modal-footer").click(); await w(900);
  out.notice = $(".alert-success")?.textContent; out.before = before; out.after = status();
  out.gone = (await fetch("/api/entries/" + id)).status;
  return out; })()`);
check("entries: delete asks first, then deletes and counts", /^Delete entry #\d+\? This cannot be undone\.$/.test(r.confirm) && r.notice.includes("deleted") && r.before === "41 found" && r.after === "40 found" && r.gone === 400, JSON.stringify(r));
check("entries: no errors in the console", problems(page.logs).length === 0, problems(page.logs).join(" | "));

// --- Theme: kept in the browser ------------------------------------------------

console.log("--- theme");
await page.goto("/about");
r = await page.evaluate(`(async () => { ${HELPERS}
  const select = $("#island-theme select");
  const start = select.value;
  type(select, "dark", "change"); await w(300);
  return { start, theme: document.documentElement.dataset.bsTheme, stored: localStorage.getItem("theme") }; })()`);
check("choosing dark applies it and remembers it", r.start === "auto" && r.theme === "dark" && r.stored === "dark", JSON.stringify(r));
await page.goto("/about");
r = await page.evaluate(`({ theme: document.documentElement.dataset.bsTheme, shows: document.querySelector("#island-theme select").value })`);
check("a new page load restores it, without a mismatch", r.theme === "dark" && r.shows === "dark" && problems(page.logs).length === 0, JSON.stringify([r, problems(page.logs)]));

page.close();
console.log(failures === 0 ? "\nall good" : `\n${failures} failed`);
process.exit(failures === 0 ? 0 : 1);
