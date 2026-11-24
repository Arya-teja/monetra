import type { CashflowType } from "@/features/categories/types";

export type Transaction = {
  id: string;
  accountId: string;
  accountName: string;
  categoryId: string;
  categoryName: string;
  type: CashflowType;
  amount: number;
  occurredAt: string; // YYYY-MM-DD
  note: string | null;
  createdAt: string;
};

// Filters read from the URL of /transactions; undefined means "no filter".
export type TransactionFilters = {
  month?: string; // YYYY-MM
  accountId?: string;
  categoryId?: string;
  type?: CashflowType;
  page: number;
};
