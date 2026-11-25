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

test("too many wrong passwords lock the login with a clear message", async ({ page }) => {
  const email = uniqueEmail("lockout");
  await createUser(email);

  // Waits for each login request, so a message left over from the previous attempt
  // cannot pass for the result of this one.
  async function submit(password: string) {
    await page.getByLabel("Password").fill(password);
    await Promise.all([
      page.waitForResponse((res) => res.url().includes("/api/auth/callback/credentials")),
      page.getByRole("button", { name: "Login" }).click(),
    ]);
  }

  await page.goto("/login");
  await page.getByLabel("Email").fill(email);
  for (let attempt = 1; attempt <= 5; attempt++) {
    await submit(`salah-${attempt}`);
    await expect(page.getByText("Email atau password salah")).toBeVisible();
    // A failed login keeps the email, so the user only retypes the password.
    await expect(page.getByLabel("Email")).toHaveValue(email);
  }

  // Locked now: even the right password is refused with a wait time.
  await submit(PASSWORD);
  await expect(page.getByText("Terlalu banyak percobaan login. Coba lagi dalam 15 menit.")).toBeVisible();
  await expect(page).toHaveURL(/\/login$/);
});
