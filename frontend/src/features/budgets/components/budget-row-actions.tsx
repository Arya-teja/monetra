"use client";

import { useState } from "react";
import { deleteBudget, updateBudgetAmount } from "../actions";
import type { Budget } from "../types";
import { Button } from "@/components/ui/button";
import { RupiahInput } from "@/components/rupiah-input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { PencilSimpleIcon, TrashIcon } from "@phosphor-icons/react";

export function BudgetRowActions({ budget }: { budget: Budget }) {
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleUpdate(formData: FormData) {
    setLoading(true);
    setError(null);
    const result = await updateBudgetAmount(budget.id, formData);
    setLoading(false);
    if ("error" in result) {
      setError(result.error);
      return;
    }
    setOpen(false);
  }

  async function handleDelete() {
    if (!window.confirm(`Hapus budget ${budget.categoryName}?`)) return;

    setLoading(true);
    const result = await deleteBudget(budget.id);
    setLoading(false);
    if ("error" in result) {
      window.alert(result.error);
    }
  }

  return (
    <div className="flex items-center gap-1">
      <Dialog
        open={open}
        onOpenChange={(value) => {
          setOpen(value);
          if (!value) setError(null);
        }}
      >
        <DialogTrigger asChild>
          <Button variant="ghost" size="icon" aria-label={`Ubah budget ${budget.categoryName}`}>
            <PencilSimpleIcon className="size-4" />
          </Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Ubah Budget {budget.categoryName}</DialogTitle>
          </DialogHeader>
          <form action={handleUpdate} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor={`budget-amount-${budget.id}`}>Nominal Budget</Label>
              <RupiahInput
                id={`budget-amount-${budget.id}`}
                name="amount"
                required
                defaultValue={budget.amount}
              />
            </div>
            {error && (
              <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {error}
              </p>
            )}
            <Button type="submit" disabled={loading} className="h-11 w-full">
              {loading ? "Menyimpan..." : "Simpan"}
            </Button>
          </form>
        </DialogContent>
      </Dialog>

      <Button
        variant="ghost"
        size="icon"
        onClick={handleDelete}
        disabled={loading}
        aria-label={`Hapus budget ${budget.categoryName}`}
      >
        <TrashIcon className="size-4 text-destructive" />
      </Button>
    </div>
  );
}
