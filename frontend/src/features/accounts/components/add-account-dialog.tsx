"use client";

import { useState } from "react";
import { createAccount } from "../actions";
import { accountTypeLabels, accountTypes } from "../types";
import { Button } from "@/components/ui/button";
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
import { PlusIcon } from "@phosphor-icons/react";

export function AddAccountDialog() {
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(formData: FormData) {
    setLoading(true);
    setError(null);

    const result = await createAccount(formData);

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
        <Button size="sm">
          <PlusIcon className="mr-1 size-4" />
          Tambah Akun
        </Button>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Tambah Akun</DialogTitle>
        </DialogHeader>

        <form action={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Nama Akun</Label>
            <Input
              id="name"
              name="name"
              required
              maxLength={50}
              placeholder="Contoh: Tabungan Utama"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="type">Tipe</Label>
            <Select name="type" defaultValue="bank">
              <SelectTrigger id="type" className="w-full">
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
            <Label htmlFor="institutionName">Nama Institusi (opsional)</Label>
            <Input
              id="institutionName"
              name="institutionName"
              maxLength={50}
              placeholder="Contoh: BSI, GoPay"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="accountNumber">Nomor Akun (opsional)</Label>
            <Input
              id="accountNumber"
              name="accountNumber"
              maxLength={50}
              placeholder="Nomor rekening/akun"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="balance">Saldo Awal</Label>
            <RupiahInput id="balance" name="balance" placeholder="0" />
            <p className="text-xs text-muted-foreground">Kosongkan kalau saldonya Rp0.</p>
          </div>

          {error && (
            <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </p>
          )}

          <Button type="submit" disabled={loading} className="w-full">
            {loading ? "Menyimpan..." : "Simpan"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
