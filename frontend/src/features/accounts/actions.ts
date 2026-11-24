"use server";

import { revalidatePath } from "next/cache";
import { apiRequest, type ApiResult } from "@/lib/api";
import { getAccessToken } from "@/lib/session";
import {
  createAccountSchema,
  updateAccountSchema,
} from "@/features/accounts/schemas";
import type { Account } from "@/features/accounts/types";

type ActionResult = { success: true } | { error: string };

function readAccountFields(data: FormData) {
  return {
    name: data.get("name"),
    type: data.get("type"),
    institutionName: data.get("institutionName"),
    accountNumber: data.get("accountNumber"),
  };
}

function toErrorMessage(code: string): string {
  switch (code) {
    case "NOT_FOUND":
      return "Akun tidak ditemukan.";
    case "CONFLICT":
      return "Akun ini sudah punya transaksi, jadi tidak bisa dihapus. Arsipkan saja.";
    default:
      return "Terjadi kesalahan. Silakan coba lagi.";
  }
}

export async function getAccounts(): Promise<ApiResult<Account[]>> {
  return apiRequest<Account[]>("/accounts?includeArchived=true", {
    method: "GET",
    token: await getAccessToken(),
  });
}

export async function createAccount(data: FormData): Promise<ActionResult> {
  const parsed = createAccountSchema.safeParse({
    ...readAccountFields(data),
    balance: data.get("balance"),
  });
  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  const result = await apiRequest<Account>("/accounts", {
    method: "POST",
    body: parsed.data,
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: toErrorMessage(result.error.code) };
  }

  revalidatePath("/accounts");
  return { success: true };
}

export async function updateAccount(
  id: string,
  data: FormData,
): Promise<ActionResult> {
  const parsed = updateAccountSchema.safeParse(readAccountFields(data));
  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  const result = await apiRequest<Account>(`/accounts/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: parsed.data,
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: toErrorMessage(result.error.code) };
  }

  revalidatePath("/accounts");
  return { success: true };
}

export async function setAccountArchived(
  id: string,
  isArchived: boolean,
): Promise<ActionResult> {
  const result = await apiRequest<Account>(`/accounts/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: { isArchived },
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: toErrorMessage(result.error.code) };
  }

  revalidatePath("/accounts");
  return { success: true };
}

export async function deleteAccount(id: string): Promise<ActionResult> {
  const result = await apiRequest<void>(`/accounts/${encodeURIComponent(id)}`, {
    method: "DELETE",
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: toErrorMessage(result.error.code) };
  }

  revalidatePath("/accounts");
  return { success: true };
}
