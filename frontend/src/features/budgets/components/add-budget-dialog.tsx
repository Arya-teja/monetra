"use client";

import { useState } from "react";
import { createBudget } from "../actions";
import type { Category } from "@/features/categories/types";
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
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(formData: FormData) {
    setLoading(true);
    setError(null);
    const result = await createBudget(formData);
    setLoading(false);
    if ("error" in result) {
      setError(result.error);
      return;
    }
    setOpen(false);
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(value) => {
        setOpen(value);
        if (!value) setError(null);
      }}
    >
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
  );
}
