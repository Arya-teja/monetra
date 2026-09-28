import { getCategories } from "@/features/categories/actions";
import { AddCategoryDialog } from "@/features/categories/components/add-category-dialog";
import { CategoryRowActions } from "@/features/categories/components/category-row-actions";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export default async function CategoriesPage() {
  const categories = await getCategories();

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-heading text-2xl font-semibold">Kategori</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Kelompokkan pemasukan dan pengeluaran kamu, misalnya Gaji, Makanan,
            atau Transport.
          </p>
        </div>
        <AddCategoryDialog />
      </div>

      <div className="mt-6 rounded-lg border">
        {categories.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-16 text-center">
            <p className="text-sm text-muted-foreground">
              Kamu belum punya kategori.
            </p>
            <p className="text-xs text-muted-foreground">
              Tambahkan kategori pertama kamu sebelum mencatat transaksi.
            </p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nama Kategori</TableHead>
                <TableHead>Tipe</TableHead>
                <TableHead className="w-24"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {categories.map((category) => (
                <TableRow key={category.id}>
                  <TableCell className="font-medium">{category.name}</TableCell>
                  <TableCell>
                    <span
                      className={
                        category.type === "INCOME"
                          ? "rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs font-medium text-emerald-600"
                          : "rounded-full bg-destructive/10 px-2 py-0.5 text-xs font-medium text-destructive"
                      }
                    >
                      {category.type === "INCOME" ? "Pemasukan" : "Pengeluaran"}
                    </span>
                  </TableCell>
                  <TableCell>
                    <CategoryRowActions category={category} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </div>
    </div>
  );
}
