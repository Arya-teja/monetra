"use client";

import { useState } from "react";
import { deleteAccount, setAccountArchived } from "../actions";
import type { Account } from "../types";
import { Button } from "@/components/ui/button";
import {
  ArchiveIcon,
  ArrowCounterClockwiseIcon,
  TrashIcon,
} from "@phosphor-icons/react";
import { EditAccountDialog } from "./edit-account-dialog";

export function AccountRowActions({ account }: { account: Account }) {
  const [loading, setLoading] = useState(false);

  async function handleToggleArchive() {
    setLoading(true);
    const result = await setAccountArchived(account.id, !account.isArchived);
    setLoading(false);
    if ("error" in result) {
      window.alert(result.error);
    }
  }

  async function handleDelete() {
    const confirmed = window.confirm(`Yakin mau hapus akun "${account.name}"?`);
    if (!confirmed) return;

    setLoading(true);
    const result = await deleteAccount(account.id);
    setLoading(false);
    if ("error" in result) {
      window.alert(result.error);
    }
  }

  return (
    <div className="flex items-center justify-end gap-1">
      <EditAccountDialog account={account} />
      <Button
        variant="ghost"
        size="icon"
        onClick={handleToggleArchive}
        disabled={loading}
        aria-label={
          account.isArchived
            ? `Aktifkan akun ${account.name}`
            : `Arsipkan akun ${account.name}`
        }
      >
        {account.isArchived ? (
          <ArrowCounterClockwiseIcon className="size-4" />
        ) : (
          <ArchiveIcon className="size-4" />
        )}
      </Button>
      <Button
        variant="ghost"
        size="icon"
        onClick={handleDelete}
        disabled={loading}
        aria-label={`Hapus akun ${account.name}`}
      >
        <TrashIcon className="size-4 text-destructive" />
      </Button>
    </div>
  );
}
