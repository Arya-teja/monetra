"use server";

import { revalidatePath } from "next/cache";
import { apiRequest, type ApiResult } from "@/lib/api";
import { getAccessToken } from "@/lib/session";
import { transactionSchema } from "@/features/transactions/schemas";
import type {
  Transaction,
  TransactionFilters,
} from "@/features/transactions/types";
import type { Budget } from "@/features/budgets/types";

type ActionResult = { success: true } | { error: string };
type SaveResult = { success: true; budgetWarning?: string } | { error: string };

function readTransactionFields(data: FormData) {
  return {
    type: data.get("type"),
    accountId: data.get("accountId"),
    categoryId: data.get("categoryId"),
    amount: data.get("amount"),
    occurredAt: data.get("occurredAt"),
    note: data.get("note") ?? "",
  };
}

function toErrorMessage(error: { code: string; details?: Record<string, string> }): string {
  if (error.details?.accountId === "account is archived") {
    return "Akun ini sudah diarsipkan. Pilih akun lain.";
  }
  switch (error.code) {
    case "VALIDATION_ERROR":
      return "Data transaksi tidak valid. Periksa kembali isiannya.";
    case "NOT_FOUND":
      return "Transaksi tidak ditemukan.";
    case "CONFLICT":
      return "Transaksi ini baru saja diubah. Muat ulang halaman lalu coba lagi.";
    default:
      return "Terjadi kesalahan. Silakan coba lagi.";
  }
}

// Balances and budget spending change with every transaction, so these pages must refresh.
function revalidateTransactionPages() {
  revalidatePath("/transactions");
  revalidatePath("/accounts");
  revalidatePath("/budgets");
}

// Returns a warning when an expense pushes its category budget to 80% or more (PRD F5).
async function budgetWarning(t: Transaction, token: string): Promise<string | undefined> {
  if (t.type !== "expense") return undefined;

  const result = await apiRequest<Budget[]>(`/budgets?month=${t.occurredAt.slice(0, 7)}`, {
    method: "GET",
    token,
  });
  const budget = result.ok ? result.data.find((b) => b.categoryId === t.categoryId) : undefined;
  if (!budget || budget.status === "safe") return undefined;

  return budget.status === "over"
    ? `Budget ${budget.categoryName} sudah terlampaui (${budget.percentage}%).`
    : `Budget ${budget.categoryName} sudah terpakai ${budget.percentage}%.`;
}

export async function getTransactions(
  filters: TransactionFilters,
): Promise<ApiResult<Transaction[]>> {
  const query = new URLSearchParams({ page: String(filters.page) });
  if (filters.month) query.set("month", filters.month);
  if (filters.accountId) query.set("accountId", filters.accountId);
  if (filters.categoryId) query.set("categoryId", filters.categoryId);
  if (filters.type) query.set("type", filters.type);

  return apiRequest<Transaction[]>(`/transactions?${query}`, {
    method: "GET",
    token: await getAccessToken(),
  });
}

// Shared by create and update: validate, save through the API, then refresh pages and
// check whether the saved expense pushed its budget to a warning.
async function saveTransaction(
  path: string,
  method: "POST" | "PATCH",
  data: FormData,
): Promise<SaveResult> {
  const parsed = transactionSchema.safeParse(readTransactionFields(data));
  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  const token = await getAccessToken();
  const result = await apiRequest<Transaction>(path, { method, body: parsed.data, token });
  if (!result.ok) {
    return { error: toErrorMessage(result.error) };
  }

  revalidateTransactionPages();
  return { success: true, budgetWarning: await budgetWarning(result.data, token) };
}

export async function createTransaction(data: FormData): Promise<SaveResult> {
  return saveTransaction("/transactions", "POST", data);
}

export async function updateTransaction(id: string, data: FormData): Promise<SaveResult> {
  return saveTransaction(`/transactions/${encodeURIComponent(id)}`, "PATCH", data);
}

export async function deleteTransaction(id: string): Promise<ActionResult> {
  const result = await apiRequest<void>(`/transactions/${encodeURIComponent(id)}`, {
    method: "DELETE",
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: toErrorMessage(result.error) };
  }

  revalidateTransactionPages();
  return { success: true };
}
