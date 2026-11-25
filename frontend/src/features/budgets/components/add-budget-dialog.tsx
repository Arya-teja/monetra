"use client";

import { createBudget } from "../actions";
import { useFormDialog } from "@/hooks/use-form-dialog";
import type { Category } from "@/features/categories/types";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PlusIcon } from "@phosphor-icons/react";

type Props = {
  month: string; // YYYY-MM
  /** Expense categories that have no budget in this month yet. */
  categories: Category[];
};

export function AddBudgetDialog({ month, categories }: Props) {
  const { open, onOpenChange, error, loading, handleSubmit } = useFormDialog(createBudget);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>
        <Button size="sm" disabled={categories.length === 0}>
          <PlusIcon className="mr-1 size-4" />
          Tambah Budget
        </Button>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Tambah Budget</DialogTitle>
        </DialogHeader>

        <form action={handleSubmit} className="space-y-4">
          <input type="hidden" name="month" value={month} />

          <div className="space-y-2">
            <Label htmlFor="budget-category">Kategori Pengeluaran</Label>
            <Select name="categoryId" defaultValue={categories[0]?.id}>
              <SelectTrigger id="budget-category" className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {categories.map((c) => (
                  <SelectItem key={c.id} value={c.id}>
                    {c.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="budget-amount">Nominal Budget</Label>
            <RupiahInput id="budget-amount" name="amount" required placeholder="0" />
          </div>

          <FormError message={error} />

          <Button type="submit" disabled={loading} className="h-11 w-full">
            {loading ? "Menyimpan..." : "Simpan"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
