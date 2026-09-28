import type {BankAccount as PrismaBankAccount} from "@prisma/client";

export type BankAccountView = Omit<PrismaBankAccount, "balance"> & {
    balance: number;
}