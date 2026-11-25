"use client";

import { useState } from "react";
import { deleteCategory, renameCategory } from "../actions";
import { useFormDialog } from "@/hooks/use-form-dialog";
import type { Category } from "../types";
import { Button } from "@/components/ui/button";
import { FormError } from "@/components/form-error";
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
  const { open, onOpenChange, error, loading, handleSubmit } = useFormDialog((formData) =>
    renameCategory(category.id, formData),
  );
  const [deleting, setDeleting] = useState(false);

  async function handleDelete() {
    if (!window.confirm(`Hapus kategori "${category.name}"?`)) return;

    setDeleting(true);
    const result = await deleteCategory(category.id);
    setDeleting(false);
    if ("error" in result) {
      window.alert(result.error);
    }
  }

  return (
    <div className="flex items-center justify-end gap-1">
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogTrigger asChild>
          <Button variant="ghost" size="icon" aria-label={`Ubah nama ${category.name}`}>
            <PencilSimpleIcon className="size-4" />
          </Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Ubah Nama Kategori</DialogTitle>
          </DialogHeader>
          <form action={handleSubmit} className="space-y-4">
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
            <FormError message={error} />
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
        disabled={deleting}
        aria-label={`Hapus kategori ${category.name}`}
      >
        <TrashIcon className="size-4 text-destructive" />
      </Button>
    </div>
  );
}
