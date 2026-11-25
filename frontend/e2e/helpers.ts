import { expect, type Page } from "@playwright/test";
import { E2E_API_URL } from "./env";

export const PASSWORD = "rahasia123";

export function uniqueEmail(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}@example.com`;
}

async function api<T>(path: string, init: { method: string; token?: string; body?: unknown }): Promise<T> {
  const res = await fetch(`${E2E_API_URL}/api/v1${path}`, {
    method: init.method,
    headers: {
      "Content-Type": "application/json",
      ...(init.token && { Authorization: `Bearer ${init.token}` }),
    },
    body: init.body === undefined ? undefined : JSON.stringify(init.body),
  });
  const json = await res.json();
  if (!res.ok) throw new Error(`${init.method} ${path}: ${res.status} ${JSON.stringify(json)}`);
  return json.data as T;
}

// Creates a user through the API (fast test setup) and returns an API token for further setup.
export async function createUser(email: string): Promise<string> {
  await api("/auth/register", { method: "POST", body: { name: "E2E User", email, password: PASSWORD } });
  const { accessToken } = await api<{ accessToken: string }>("/auth/login", {
    method: "POST",
    body: { email, password: PASSWORD },
  });
  return accessToken;
}

export async function createAccount(token: string, name: string, balance: number) {
  return api<{ id: string }>("/accounts", { method: "POST", token, body: { name, type: "bank", balance } });
}

async function categoryId(token: string, name: string) {
  const categories = await api<{ id: string; name: string }[]>("/categories", { method: "GET", token });
  const category = categories.find((c) => c.name === name);
  if (!category) throw new Error(`category ${name} not found`);
  return category.id;
}

export async function createBudget(token: string, categoryName: string, month: string, amount: number) {
  const id = await categoryId(token, categoryName);
  return api("/budgets", { method: "POST", token, body: { categoryId: id, month, amount } });
}

export async function login(page: Page, email: string) {
  await page.goto("/login");
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill(PASSWORD);
  await page.getByRole("button", { name: "Login" }).click();
  // Login redirects and renders the full dashboard; WebKit under parallel test load
  // can need more than the default 5 s for that.
  await expect(page).toHaveURL(/\/dashboard$/, { timeout: 15_000 });
}

// Formats like the app does (Intl id-ID uses a non-breaking space after "Rp").
export function rupiah(amount: number) {
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(amount);
}

// Current month in Asia/Jakarta as YYYY-MM, matching the app's default month.
export function currentMonth() {
  return new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Jakarta" }).format(new Date()).slice(0, 7);
}

// Logs the user out through the API, as another device would: revokes every token.
export async function logoutEverywhere(token: string) {
  const res = await fetch(`${E2E_API_URL}/api/v1/auth/logout`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (res.status !== 204) throw new Error(`logout: ${res.status}`);
}
