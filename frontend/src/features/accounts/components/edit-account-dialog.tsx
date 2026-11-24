"use client";

import { useState } from "react";
import { updateAccount } from "../actions";
import { accountTypeLabels, accountTypes, type Account } from "../types";
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
import { PencilSimpleIcon } from "@phosphor-icons/react";

export function EditAccountDialog({ account }: { account: Account }) {
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(formData: FormData) {
    setLoading(true);
    setError(null);

    const result = await updateAccount(account.id, formData);

    if ("error" in result) {
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
        <Button variant="ghost" size="icon" aria-label={`Edit akun ${account.name}`}>
          <PencilSimpleIcon className="size-4" />
        </Button>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Akun</DialogTitle>
        </DialogHeader>

        <form action={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-name">Nama Akun</Label>
            <Input
              id="edit-name"
              name="name"
              required
              maxLength={50}
              defaultValue={account.name}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-type">Tipe</Label>
            <Select name="type" defaultValue={account.type}>
              <SelectTrigger id="edit-type" className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {accountTypes.map((type) => (
                  <SelectItem key={type} value={type}>
                    {accountTypeLabels[type]}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-institutionName">Nama Institusi (opsional)</Label>
            <Input
              id="edit-institutionName"
              name="institutionName"
              maxLength={50}
              defaultValue={account.institutionName ?? ""}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-accountNumber">Nomor Akun (opsional)</Label>
            <Input
              id="edit-accountNumber"
              name="accountNumber"
              maxLength={50}
              defaultValue={account.accountNumber ?? ""}
            />
          </div>

          <p className="text-xs text-muted-foreground">
            Saldo tidak bisa diubah di sini. Saldo hanya berubah lewat transaksi.
          </p>

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
