"use client";

import { useState } from "react";
import { toast } from "sonner";
import { createTransaction, updateTransaction } from "../actions";
import type { Transaction } from "../types";
import { createCategory } from "@/features/categories/actions";
import {
  cashflowTypeLabels,
  cashflowTypes,
  type CashflowType,
  type Category,
} from "@/features/categories/types";
import type { Account } from "@/features/accounts/types";
import { todayInJakarta } from "@/lib/format";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { FormError } from "@/components/form-error";
import { Input } from "@/components/ui/input";
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

const ADD_CATEGORY = "__add_category__";

type Props = {
  accounts: Account[];
  categories: Category[];
  /** When set, the dialog edits this transaction; otherwise it creates one. */
  transaction?: Transaction;
  /** Element that opens the dialog, e.g. the (+) button of the navigation. */
  trigger: React.ReactNode;
};

export function TransactionFormDialog({ accounts, categories, transaction, trigger }: Props) {
  const idPrefix = `tx-${transaction?.id ?? "new"}`;
  const [open, setOpen] = useState(false);
  const [type, setType] = useState<CashflowType>(transaction?.type ?? "expense");
  const [categoryId, setCategoryId] = useState(transaction?.categoryId ?? "");
  const [addedCategories, setAddedCategories] = useState<Category[]>([]);
  const [newCategoryName, setNewCategoryName] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // Archived accounts are hidden, except the one an edited transaction already uses.
  const selectableAccounts = accounts.filter(
    (a) => !a.isArchived || a.id === transaction?.accountId,
  );
  const categoryOptions = [
    ...categories,
    ...addedCategories.filter((added) => !categories.some((c) => c.id === added.id)),
  ].filter((c) => c.type === type);

  function resetForm() {
    setType(transaction?.type ?? "expense");
    setCategoryId(transaction?.categoryId ?? "");
    setNewCategoryName(null);
    setError(null);
  }

  function handleTypeChange(next: CashflowType) {
    setType(next);
    setCategoryId("");
  }

  async function handleAddCategory() {
    if (newCategoryName === null) return;
    setLoading(true);
    setError(null);
    const result = await createCategory(newCategoryName, type);
    setLoading(false);
    if ("error" in result) {
      setError(result.error);
      return;
    }
    setAddedCategories((prev) => [...prev, result.category]);
    setCategoryId(result.category.id);
    setNewCategoryName(null);
  }

  async function handleSubmit(formData: FormData) {
    setLoading(true);
    setError(null);
    const result = transaction
      ? await updateTransaction(transaction.id, formData)
      : await createTransaction(formData);
    setLoading(false);
    if ("error" in result) {
      setError(result.error);
      return;
    }
    setOpen(false);
    if (result.budgetWarning) {
      toast.warning(result.budgetWarning);
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(value) => {
        if (value) resetForm();
        setOpen(value);
      }}
    >
      <DialogTrigger asChild>{trigger}</DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>{transaction ? "Edit Transaksi" : "Tambah Transaksi"}</DialogTitle>
        </DialogHeader>

        {selectableAccounts.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Kamu belum punya akun aktif. Tambahkan akun dulu di halaman Akun.
          </p>
        ) : (
          <form action={handleSubmit} className="space-y-4">
            <fieldset>
              <legend className="sr-only">Tipe transaksi</legend>
              <div className="grid grid-cols-2 gap-1 rounded-lg bg-muted p-1">
                {cashflowTypes.map((t) => (
                  <label
                    key={t}
                    className={cn(
                      "flex h-11 cursor-pointer items-center justify-center rounded-md text-sm font-medium has-focus-visible:ring-2 has-focus-visible:ring-ring",
                      type === t ? "bg-background shadow-sm" : "text-muted-foreground",
                    )}
                  >
                    <input
                      type="radio"
                      name="type"
                      value={t}
                      checked={type === t}
                      onChange={() => handleTypeChange(t)}
                      className="sr-only"
                    />
                    {cashflowTypeLabels[t]}
                  </label>
                ))}
              </div>
            </fieldset>

            <div className="space-y-2">
              <Label htmlFor={`${idPrefix}-account`}>Akun</Label>
              <Select
                name="accountId"
                defaultValue={transaction?.accountId ?? selectableAccounts[0].id}
              >
                <SelectTrigger id={`${idPrefix}-account`} className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {selectableAccounts.map((a) => (
                    <SelectItem key={a.id} value={a.id}>
                      {a.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor={`${idPrefix}-category`}>Kategori</Label>
              <Select
                name="categoryId"
                value={categoryId}
                onValueChange={(value) =>
                  value === ADD_CATEGORY ? setNewCategoryName("") : setCategoryId(value)
                }
              >
                <SelectTrigger id={`${idPrefix}-category`} className="w-full">
                  <SelectValue placeholder="Pilih kategori" />
                </SelectTrigger>
                <SelectContent>
                  {categoryOptions.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                  <SelectItem value={ADD_CATEGORY}>+ Tambah kategori</SelectItem>
                </SelectContent>
              </Select>

              {newCategoryName !== null && (
                <div className="flex gap-2">
                  <Input
                    aria-label="Nama kategori baru"
                    placeholder="Nama kategori baru"
                    maxLength={50}
                    autoFocus
                    value={newCategoryName}
                    onChange={(e) => setNewCategoryName(e.target.value)}
                  />
                  <Button type="button" onClick={handleAddCategory} disabled={loading}>
                    Tambah
                  </Button>
                  <Button type="button" variant="ghost" onClick={() => setNewCategoryName(null)}>
                    Batal
                  </Button>
                </div>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor={`${idPrefix}-amount`}>Nominal</Label>
              <RupiahInput
                id={`${idPrefix}-amount`}
                name="amount"
                required
                placeholder="0"
                defaultValue={transaction?.amount}
                className="h-14 text-2xl font-semibold"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor={`${idPrefix}-date`}>Tanggal</Label>
              <Input
                id={`${idPrefix}-date`}
                name="occurredAt"
                type="date"
                required
                defaultValue={transaction?.occurredAt ?? todayInJakarta()}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor={`${idPrefix}-note`}>Catatan (opsional)</Label>
              <Input
                id={`${idPrefix}-note`}
                name="note"
                maxLength={200}
                defaultValue={transaction?.note ?? ""}
              />
            </div>

            <FormError message={error} />

            <Button type="submit" disabled={loading} className="h-11 w-full">
              {loading ? "Menyimpan..." : "Simpan"}
            </Button>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
