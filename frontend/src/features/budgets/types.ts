export type BudgetStatus = "safe" | "warning" | "over";

export const budgetStatusLabels: Record<BudgetStatus, string> = {
  safe: "Aman",
  warning: "Hampir habis",
  over: "Over budget",
};

export type Budget = {
  id: string;
  categoryId: string;
  categoryName: string;
  month: string; // YYYY-MM
  amount: number;
  spent: number;
  percentage: number;
  status: BudgetStatus;
};
