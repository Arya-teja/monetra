import type { Budget } from "@/features/budgets/types";
import type { Transaction } from "@/features/transactions/types";

export type CategoryTotal = {
  categoryId: string;
  categoryName: string;
  total: number;
};

export type DashboardSummary = {
  month: string; // YYYY-MM
  totalBalance: number;
  income: number;
  expense: number;
  /** expense / income * 100, rounded down; null when there is no income. */
  usedPercentage: number | null;
  topCategories: CategoryTotal[];
  budgetAlerts: Budget[];
  recentTransactions: Transaction[];
};

export type MonthTotal = {
  month: string; // YYYY-MM
  income: number;
  expense: number;
};
