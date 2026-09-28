"use client";

import { useState } from "react";
import { updateBankAccount } from "../actions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import { Pencil } from "lucide-react";
import type { BankAccountView } from "../types/types";

export function EditAccountDialog({ account }: { account: BankAccountView }) {
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(formData: FormData) {
    setLoading(true);
    setError(null);

    const result = await updateBankAccount(account.id, formData);

    if (result?.error) {
      setError(result.error);
      setLoading(false);
      return;
    }

    setLoading(false);
    setOpen(false);
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="ghost" size="icon">
          <Pencil className="size-4" />
        </Button>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Akun Bank</DialogTitle>
        </DialogHeader>

        <form action={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-name">Nama Akun</Label>
            <Input
              id="edit-name"
              name="name"
              required
              defaultValue={account.name}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-bankName">Nama Bank</Label>
            <Input
              id="edit-bankName"
              name="bankName"
              required
              defaultValue={account.bankName}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-accountNumber">Nomor Akun</Label>
            <Input
              id="edit-accountNumber"
              name="accountNumber"
              required
              defaultValue={account.accountNumber}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-type">Tipe</Label>
            <Select name="type" defaultValue={account.type}>
              <SelectTrigger id="edit-type" className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="BANK">Bank</SelectItem>
                <SelectItem value="EWALLET">E-Wallet</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-balance">Saldo</Label>
            <Input
              id="edit-balance"
              name="balance"
              type="number"
              min="0"
              step="1000"
              required
              defaultValue={account.balance}
            />
          </div>

          {error && (
            <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </p>
          )}

          <Button type="submit" disabled={loading} className="w-full">
            {loading ? "Menyimpan..." : "Simpan Perubahan"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}