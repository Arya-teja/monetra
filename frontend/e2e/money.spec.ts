import { expect, test, type Page } from "@playwright/test";
import { createAccount, createBudget, createUser, currentMonth, login, rupiah, uniqueEmail } from "./helpers";

async function chooseOption(page: Page, label: string, option: string) {
  await page.getByLabel(label).click();
  await page.getByRole("option", { name: option }).click();
}

// Opens the transaction form from the global (+) button (bottom bar on phones, sidebar on desktop).
async function addExpense(page: Page, category: string, amount: string) {
  await page.getByRole("button", { name: "Tambah transaksi" }).click();
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByText("Pengeluaran")).toBeVisible();
  await chooseOption(page, "Kategori", category);
  await dialog.getByLabel("Nominal").fill(amount);
  await dialog.getByRole("button", { name: "Simpan" }).click();
  await expect(dialog).toBeHidden();
}

test("new account can start at Rp0 by leaving the starting balance empty", async ({ page }) => {
  const email = uniqueEmail("zero");
  await createUser(email);
  await login(page, email);

  await page.goto("/accounts");
  await page.getByRole("button", { name: "Tambah Akun" }).click();
  await page.getByLabel("Nama Akun").fill("Dompet");
  await page.getByRole("dialog").getByRole("button", { name: "Simpan" }).click();

  const row = page.getByRole("listitem").filter({ hasText: "Dompet" });
  await expect(row).toContainText(rupiah(0));
});

test("money inputs show thousand separators while typing", async ({ page }) => {
  const email = uniqueEmail("typing");
  const token = await createUser(email);
  await createAccount(token, "BSI", 0);
  await login(page, email);

  await page.getByRole("button", { name: "Tambah transaksi" }).click();
  const amount = page.getByRole("dialog").getByLabel("Nominal");
  await amount.pressSequentially("1500000");
  await expect(amount).toHaveValue("1.500.000");
  await amount.press("Backspace");
  await expect(amount).toHaveValue("150.000");
});

test("an expense added from the dashboard updates the dashboard and the account balance", async ({ page }) => {
  const email = uniqueEmail("expense");
  const token = await createUser(email);
  await createAccount(token, "BSI", 1_000_000);
  await login(page, email);

  await addExpense(page, "Belanja", "250000");

  // The dashboard the user is on must show the new numbers without a manual reload.
  await expect(page.getByText(rupiah(750_000)).first()).toBeVisible();
  await expect(page.getByText(rupiah(250_000)).first()).toBeVisible();

  // Navigating within the app must not show a stale balance.
  await page.goto("/accounts");
  await expect(page.getByRole("listitem").filter({ hasText: "BSI" })).toContainText(rupiah(750_000));
});

test("an expense that reaches 80% of its budget shows a warning toast", async ({ page }) => {
  const email = uniqueEmail("budget");
  const token = await createUser(email);
  await createAccount(token, "BSI", 1_000_000);
  await createBudget(token, "Hiburan", currentMonth(), 100_000);
  await login(page, email);

  await addExpense(page, "Hiburan", "85000");
  await expect(page.getByText("Budget Hiburan sudah terpakai 85%.")).toBeVisible();

  await page.goto("/budgets");
  await expect(page.getByText("Hampir habis")).toBeVisible();
});

test("transactions page filters by type through the URL", async ({ page }) => {
  const email = uniqueEmail("filter");
  const token = await createUser(email);
  await createAccount(token, "BSI", 1_000_000);
  await login(page, email);
  await addExpense(page, "Belanja", "10000");

  await page.goto("/transactions?type=income");
  await expect(page.getByText("Belum ada transaksi untuk filter ini.")).toBeVisible();
  await page.goto("/transactions?type=expense");
  // exact: the filter dropdown also contains an option "Belanja (Pengeluaran)".
  await expect(page.getByText("Belanja", { exact: true })).toBeVisible();
});
