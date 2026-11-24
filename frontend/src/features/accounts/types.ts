export const accountTypes = ["bank", "ewallet", "cash"] as const;

export type AccountType = (typeof accountTypes)[number];

export const accountTypeLabels: Record<AccountType, string> = {
  bank: "Bank",
  ewallet: "E-Wallet",
  cash: "Cash",
};

export type Account = {
  id: string;
  name: string;
  type: AccountType;
  institutionName: string | null;
  accountNumber: string | null;
  balance: number;
  isArchived: boolean;
  createdAt: string;
};
