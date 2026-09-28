"use client";

import { useState } from "react";
import { deleteBankAccount } from "../actions";
import { Button } from "@/components/ui/button";
import { Trash2 } from "lucide-react";
import type { BankAccount } from "@prisma/client";
import { EditAccountDialog } from "./edit-account-dialog";
import type { BankAccountView } from "../types/types";

export function AccountRowActions({ account }: { account: BankAccountView }) {
  const [loading, setLoading] = useState(false);

  async function handleDelete() {
    const confirmed = window.confirm(
      `Yakin mau hapus akun "${account.name}"?`
    );
    if (!confirmed) return;

    setLoading(true);
    await deleteBankAccount(account.id);
    setLoading(false);
  }

  return (
   <div className="flex items-center gap-1">
      <EditAccountDialog account={account} />
      <Button
        variant="ghost"
        size="icon"
        onClick={handleDelete}
        disabled={loading}
      >
        <Trash2 className="size-4 text-destructive" />
      </Button>
    </div>
  );

}