"use client";

import { useState } from "react";
import { deleteCategory } from "@/features/categories/actions";
import { Button } from "@/components/ui/button";
import { Trash2 } from "lucide-react";
import { EditCategoryDialog } from "@/features/categories/components/edit-category-dialog";
import type { Category } from "@prisma/client";

export function CategoryRowActions({ category }: { category: Category }) {
  const [loading, setLoading] = useState(false);

  async function handleDelete() {
    const confirmed = window.confirm(
      `Hapus kategori "${category.name}"?\n\nTransaksi yang memakai kategori ini akan menjadi tanpa kategori, dan budget untuk kategori ini ikut terhapus.`
    );
    if (!confirmed) return;

    setLoading(true);
    await deleteCategory(category.id);
    setLoading(false);
  }

  return (
    <div className="flex items-center gap-1">
      <EditCategoryDialog category={category} />
      <Button
        variant="ghost"
        size="icon"
        onClick={handleDelete}
        disabled={loading}
      >
        <Trash2 className="size-4 text-destructive" />
      </Button>
    </div>
  );
}