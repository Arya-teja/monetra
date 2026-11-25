import { readFileSync } from "node:fs";
import { expect, test } from "@playwright/test";
import { encode } from "next-auth/jwt";
import { createUser, login, logoutEverywhere, PASSWORD, uniqueEmail } from "./helpers";

test("register, then log in and out", async ({ page }) => {
  const email = uniqueEmail("register");

  await page.goto("/register");
  await page.getByLabel("Nama").fill("Arya E2E");
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill(PASSWORD);
  await page.getByRole("button", { name: "Register" }).click();
  await expect(page).toHaveURL(/\/login$/);

  await login(page, email);
  await expect(page.getByRole("heading", { name: "Halo, Arya E2E" })).toBeVisible();

  await page.getByRole("button", { name: "Logout" }).click();
  await expect(page).toHaveURL(/\/login$/);
  await page.goto("/dashboard");
  await expect(page).toHaveURL(/\/login$/);
});

test("register accepts an email with surrounding spaces", async ({ page }) => {
  // Phone keyboards often add a trailing space after autocomplete.
  await page.goto("/register");
  await page.getByLabel("Nama").fill("Spasi");
  await page.getByLabel("Email").fill(`  ${uniqueEmail("spaces")}  `);
  await page.getByLabel("Password").fill(PASSWORD);
  await page.getByRole("button", { name: "Register" }).click();
  await expect(page).toHaveURL(/\/login$/);
});

test("wrong password shows a clear message", async ({ page }) => {
  const email = uniqueEmail("wrongpass");
  await createUser(email);

  await page.goto("/login");
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill("salah-salah");
  await page.getByRole("button", { name: "Login" }).click();
  await expect(page.getByText("Email atau password salah")).toBeVisible();
});

test("a session cookie from before the Go API redirects to login once (no loop)", async ({ page, context, baseURL }) => {
  const secret = readFileSync(".env", "utf8").match(/^AUTH_SECRET=["']?([^"'\r\n]+)/m)?.[1];
  if (!secret) throw new Error("AUTH_SECRET missing in frontend/.env");

  // Shape of a Stage 0 (Prisma era) session: a user but no Go access token.
  const legacy = await encode({
    token: { name: "Legacy", email: "legacy@example.com", sub: "legacy-cuid" },
    secret,
    salt: "authjs.session-token",
  });
  await context.addCookies([{ name: "authjs.session-token", value: legacy, url: baseURL! }]);

  await page.goto("/dashboard");
  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByRole("button", { name: "Login" })).toBeVisible();
});

test("logging out on another device ends this session on the next page", async ({ page }) => {
  const email = uniqueEmail("elsewhere");
  const otherDevice = await createUser(email);
  await login(page, email);

  await logoutEverywhere(otherDevice);

  await page.goto("/accounts");
  await expect(page).toHaveURL(/\/login$/);
  await page.goto("/dashboard");
  await expect(page).toHaveURL(/\/login$/);
});
