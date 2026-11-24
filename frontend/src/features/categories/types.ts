export const cashflowTypes = ["income", "expense"] as const;

export type CashflowType = (typeof cashflowTypes)[number];

export const cashflowTypeLabels: Record<CashflowType, string> = {
  income: "Pemasukan",
  expense: "Pengeluaran",
};

export type Category = {
  id: string;
  name: string;
  type: CashflowType;
  isDefault: boolean;
};
