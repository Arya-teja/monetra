"use client";

import { useState } from "react";
import { deleteCategory, renameCategory } from "../actions";
import type { Category } from "../types";
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
import { PencilSimpleIcon, TrashIcon } from "@phosphor-icons/react";

export function CategoryRowActions({ category }: { category: Category }) {
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleRename(formData: FormData) {
    setLoading(true);
    setError(null);
    const result = await renameCategory(category.id, formData);
    setLoading(false);
    if ("error" in result) {
      setError(result.error);
      return;
    }
    setOpen(false);
  }

  async function handleDelete() {
    if (!window.confirm(`Hapus kategori "${category.name}"?`)) return;

    setLoading(true);
    const result = await deleteCategory(category.id);
    setLoading(false);
    if ("error" in result) {
      window.alert(result.error);
    }
  }

  return (
    <div className="flex items-center justify-end gap-1">
      <Dialog
        open={open}
        onOpenChange={(value) => {
          setOpen(value);
          if (!value) setError(null);
        }}
      >
        <DialogTrigger asChild>
          <Button variant="ghost" size="icon" aria-label={`Ubah nama ${category.name}`}>
            <PencilSimpleIcon className="size-4" />
          </Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Ubah Nama Kategori</DialogTitle>
          </DialogHeader>
          <form action={handleRename} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor={`category-name-${category.id}`}>Nama Kategori</Label>
              <Input
                id={`category-name-${category.id}`}
                name="name"
                required
                maxLength={50}
                defaultValue={category.name}
              />
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

      <Button
        variant="ghost"
        size="icon"
        onClick={handleDelete}
        disabled={loading}
        aria-label={`Hapus kategori ${category.name}`}
      >
        <TrashIcon className="size-4 text-destructive" />
      </Button>
    </div>
  );
}
