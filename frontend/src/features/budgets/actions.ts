"use server";

import { revalidatePath } from "next/cache";
import { apiRequest, type ApiResult } from "@/lib/api";
import { getAccessToken } from "@/lib/session";
import { budgetAmountSchema, createBudgetSchema } from "@/features/budgets/schemas";
import type { Budget } from "@/features/budgets/types";

type ActionResult = { success: true } | { error: string };

const GENERIC_ERROR = "Terjadi kesalahan. Silakan coba lagi.";

export async function getBudgets(month: string): Promise<ApiResult<Budget[]>> {
  return apiRequest<Budget[]>(`/budgets?month=${encodeURIComponent(month)}`, {
    method: "GET",
    token: await getAccessToken(),
  });
}

export async function createBudget(data: FormData): Promise<ActionResult> {
  const parsed = createBudgetSchema.safeParse({
    categoryId: data.get("categoryId"),
    month: data.get("month"),
    amount: data.get("amount"),
  });
  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  const result = await apiRequest<Budget>("/budgets", {
    method: "POST",
    body: parsed.data,
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return {
      error:
        result.error.code === "CONFLICT"
          ? "Kategori ini sudah punya budget di bulan ini."
          : GENERIC_ERROR,
    };
  }

  revalidatePath("/budgets");
  return { success: true };
}

export async function updateBudgetAmount(id: string, data: FormData): Promise<ActionResult> {
  const parsed = budgetAmountSchema.safeParse(data.get("amount"));
  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  const result = await apiRequest<Budget>(`/budgets/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: { amount: parsed.data },
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: result.error.code === "NOT_FOUND" ? "Budget tidak ditemukan." : GENERIC_ERROR };
  }

  revalidatePath("/budgets");
  return { success: true };
}

export async function deleteBudget(id: string): Promise<ActionResult> {
  const result = await apiRequest<void>(`/budgets/${encodeURIComponent(id)}`, {
    method: "DELETE",
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: result.error.code === "NOT_FOUND" ? "Budget tidak ditemukan." : GENERIC_ERROR };
  }

  revalidatePath("/budgets");
  return { success: true };
}

export async function copyBudgetsFromPreviousMonth(
  month: string,
): Promise<{ copied: number } | { error: string }> {
  const result = await apiRequest<{ copied: number }>("/budgets/copy", {
    method: "POST",
    body: { month },
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: GENERIC_ERROR };
  }

  revalidatePath("/budgets");
  return { copied: result.data.copied };
}
