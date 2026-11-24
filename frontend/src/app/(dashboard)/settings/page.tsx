import { getCategories } from "@/features/categories/actions";
import { CategoryRowActions } from "@/features/categories/components/category-row-actions";
import { cashflowTypeLabels, cashflowTypes } from "@/features/categories/types";

export default async function SettingsPage() {
  const result = await getCategories();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-heading text-2xl font-semibold">Pengaturan</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Ubah nama atau hapus kategori. Kategori baru dibuat langsung dari form transaksi.
        </p>
      </div>

      {!result.ok ? (
        <p className="rounded-lg border px-4 py-16 text-center text-sm text-destructive">
          Gagal memuat kategori. Silakan muat ulang halaman.
        </p>
      ) : (
        cashflowTypes.map((type) => (
          <section key={type}>
            <h2 className="mb-2 font-medium">Kategori {cashflowTypeLabels[type]}</h2>
            <ul className="divide-y rounded-lg border">
              {result.data
                .filter((c) => c.type === type)
                .map((c) => (
                  <li key={c.id} className="flex items-center gap-3 px-4 py-2">
                    <span className="flex-1">
                      {c.name}
                      {c.isDefault && (
                        <span className="ml-2 rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
                          Default
                        </span>
                      )}
                    </span>
                    <CategoryRowActions category={c} />
                  </li>
                ))}
            </ul>
          </section>
        ))
      )}
    </div>
  );
}
