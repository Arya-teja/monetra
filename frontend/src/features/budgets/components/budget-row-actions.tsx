"use client";

import { useState } from "react";
import { deleteBudget, updateBudgetAmount } from "../actions";
import { useFormDialog } from "@/hooks/use-form-dialog";
import type { Budget } from "../types";
import { Button } from "@/components/ui/button";
import { FormError } from "@/components/form-error";
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
  const { open, onOpenChange, error, loading, handleSubmit } = useFormDialog((formData) =>
    updateBudgetAmount(budget.id, formData),
  );
  const [deleting, setDeleting] = useState(false);

  async function handleDelete() {
    if (!window.confirm(`Hapus budget ${budget.categoryName}?`)) return;

    setDeleting(true);
    const result = await deleteBudget(budget.id);
    setDeleting(false);
    if ("error" in result) {
      window.alert(result.error);
    }
  }

  return (
    <div className="flex items-center gap-1">
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogTrigger asChild>
          <Button variant="ghost" size="icon" aria-label={`Ubah budget ${budget.categoryName}`}>
            <PencilSimpleIcon className="size-4" />
          </Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Ubah Budget {budget.categoryName}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor={`budget-amount-${budget.id}`}>Nominal Budget</Label>
              <RupiahInput
                id={`budget-amount-${budget.id}`}
                name="amount"
                required
                defaultValue={budget.amount}
              />
            </div>
            <FormError message={error} />
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
        disabled={deleting}
        aria-label={`Hapus budget ${budget.categoryName}`}
      >
        <TrashIcon className="size-4 text-destructive" />
      </Button>
    </div>
  );
}
