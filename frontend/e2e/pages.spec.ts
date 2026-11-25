import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { createAccount, createUser, login, uniqueEmail } from "./helpers";

const publicPages = ["/login", "/register"];
const appPages = ["/dashboard", "/transactions", "/budgets", "/accounts", "/settings"];

// Every page: no serious accessibility problems, no sideways scrolling, and a
// screenshot in test-results/screens for visual review.
async function checkPage(page: import("@playwright/test").Page, path: string, project: string) {
  await page.goto(path);
  await page.waitForLoadState("networkidle");

  const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
  expect(overflow, `${path} scrolls sideways by ${overflow}px`).toBeLessThanOrEqual(0);

  const { violations } = await new AxeBuilder({ page }).analyze();
  const serious = violations
    .filter((v) => v.impact === "serious" || v.impact === "critical")
    .map((v) => `${v.id}: ${v.help} (${v.nodes.length}x) e.g. ${v.nodes[0]?.target.join(" ")}`);
  expect(serious, `${path} accessibility`).toEqual([]);

  await page.screenshot({ path: `test-results/screens/${project}${path.replaceAll("/", "-")}.png`, fullPage: true });
}

test("public pages", async ({ page }, info) => {
  for (const path of publicPages) {
    await checkPage(page, path, info.project.name);
  }
});

test("app pages", async ({ page }, info) => {
  const email = uniqueEmail("pages");
  const token = await createUser(email);
  await createAccount(token, "BSI Tabungan", 1_500_000);
  await login(page, email);

  for (const path of appPages) {
    await checkPage(page, path, info.project.name);
  }
});
