"use client";

import { useState } from "react";
import { deleteTransaction } from "../actions";
import type { Transaction } from "../types";
import { TransactionFormDialog } from "./transaction-form-dialog";
import type { Account } from "@/features/accounts/types";
import type { Category } from "@/features/categories/types";
import { Button } from "@/components/ui/button";
import { PencilSimpleIcon, TrashIcon } from "@phosphor-icons/react";

type Props = {
  transaction: Transaction;
  accounts: Account[];
  categories: Category[];
};

export function TransactionRowActions({ transaction, accounts, categories }: Props) {
  const [loading, setLoading] = useState(false);

  async function handleDelete() {
    if (!window.confirm("Hapus transaksi ini? Saldo akun akan dikembalikan.")) return;

    setLoading(true);
    const result = await deleteTransaction(transaction.id);
    setLoading(false);
    if ("error" in result) {
      window.alert(result.error);
    }
  }

  return (
    <div className="flex items-center gap-1">
      <TransactionFormDialog
        transaction={transaction}
        accounts={accounts}
        categories={categories}
        trigger={
          <Button variant="ghost" size="icon" aria-label="Edit transaksi">
            <PencilSimpleIcon className="size-5" />
          </Button>
        }
      />
      <Button
        variant="ghost"
        size="icon"
        onClick={handleDelete}
        disabled={loading}
        aria-label="Hapus transaksi"
      >
        <TrashIcon className="size-4 text-destructive" />
      </Button>
    </div>
  );
}
