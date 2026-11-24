"use server";

import { revalidatePath } from "next/cache";
import { apiRequest, type ApiResult } from "@/lib/api";
import { getAccessToken } from "@/lib/session";
import {
  categoryNameSchema,
  createCategorySchema,
} from "@/features/categories/schemas";
import type { CashflowType, Category } from "@/features/categories/types";

const GENERIC_ERROR = "Terjadi kesalahan. Silakan coba lagi.";
const DUPLICATE_ERROR = "Kategori dengan nama ini sudah ada.";

function revalidateCategoryPages() {
  revalidatePath("/settings");
  revalidatePath("/transactions");
}

export async function getCategories(): Promise<ApiResult<Category[]>> {
  return apiRequest<Category[]>("/categories", {
    method: "GET",
    token: await getAccessToken(),
  });
}

// Used by the transaction form to add a category without leaving the form.
export async function createCategory(
  name: string,
  type: CashflowType,
): Promise<{ category: Category } | { error: string }> {
  const parsed = createCategorySchema.safeParse({ name, type });
  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  const result = await apiRequest<Category>("/categories", {
    method: "POST",
    body: parsed.data,
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: result.error.code === "CONFLICT" ? DUPLICATE_ERROR : GENERIC_ERROR };
  }

  revalidateCategoryPages();
  return { category: result.data };
}

export async function renameCategory(
  id: string,
  data: FormData,
): Promise<{ success: true } | { error: string }> {
  const parsed = categoryNameSchema.safeParse(data.get("name"));
  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  const result = await apiRequest<Category>(`/categories/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: { name: parsed.data },
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return { error: result.error.code === "CONFLICT" ? DUPLICATE_ERROR : GENERIC_ERROR };
  }

  revalidateCategoryPages();
  return { success: true };
}

export async function deleteCategory(
  id: string,
): Promise<{ success: true } | { error: string }> {
  const result = await apiRequest<void>(`/categories/${encodeURIComponent(id)}`, {
    method: "DELETE",
    token: await getAccessToken(),
  });
  if (!result.ok) {
    return {
      error:
        result.error.code === "CONFLICT"
          ? "Kategori ini sudah dipakai transaksi, jadi tidak bisa dihapus."
          : GENERIC_ERROR,
    };
  }

  revalidateCategoryPages();
  return { success: true };
}
