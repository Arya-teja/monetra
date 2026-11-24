"use server";

import { revalidatePath } from "next/cache";
import { apiRequest, type ApiResult } from "@/lib/api";
import { getAccessToken } from "@/lib/session";
import { transactionSchema } from "@/features/transactions/schemas";
import type {
  Transaction,
  TransactionFilters,
} from "@/features/transactions/types";

type ActionResult = { success: true } | { error: string };

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

// Balances change with every transaction, so both pages must refresh.
function revalidateTransactionPages() {
  revalidatePath("/transactions");
  revalidatePath("/accounts");
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

export async function createTransaction(data: FormData): Promise<ActionResult> {
  const parsed = transactionSchema.safeParse(readTransactionFields(data));
  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  const result = await apiRequest<Transaction>("/transactions", {
    method: "POST",
    body: parsed.data,
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: toErrorMessage(result.error) };
  }

  revalidateTransactionPages();
  return { success: true };
}

export async function updateTransaction(
  id: string,
  data: FormData,
): Promise<ActionResult> {
  const parsed = transactionSchema.safeParse(readTransactionFields(data));
  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  const result = await apiRequest<Transaction>(`/transactions/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: parsed.data,
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: toErrorMessage(result.error) };
  }

  revalidateTransactionPages();
  return { success: true };
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
