import Link from "next/link";
import { getBudgets } from "@/features/budgets/actions";
import { AddBudgetDialog } from "@/features/budgets/components/add-budget-dialog";
import { BudgetRowActions } from "@/features/budgets/components/budget-row-actions";
import { CopyBudgetsButton } from "@/features/budgets/components/copy-budgets-button";
import { budgetStatusLabels, type BudgetStatus } from "@/features/budgets/types";
import { getCategories } from "@/features/categories/actions";
import { Button } from "@/components/ui/button";
import { formatMonth, formatRupiah, todayInJakarta } from "@/lib/format";
import { cn } from "@/lib/utils";
import { CaretLeftIcon, CaretRightIcon } from "@phosphor-icons/react/ssr";

const statusColors: Record<BudgetStatus, { bar: string; text: string }> = {
  safe: { bar: "bg-income", text: "text-income" },
  warning: { bar: "bg-warning", text: "text-warning" },
  over: { bar: "bg-expense", text: "text-expense" },
};

// Adds delta months to "YYYY-MM".
function shiftMonth(month: string, delta: number) {
  const date = new Date(`${month}-01T00:00:00Z`);
  date.setUTCMonth(date.getUTCMonth() + delta);
  return date.toISOString().slice(0, 7);
}

export default async function BudgetsPage({ searchParams }: PageProps<"/budgets">) {
  const requested = (await searchParams).month;
  const month =
    typeof requested === "string" && /^\d{4}-\d{2}$/.test(requested)
      ? requested
      : todayInJakarta().slice(0, 7);

  const [budgetsResult, categoriesResult] = await Promise.all([
    getBudgets(month),
    getCategories(),
  ]);
  const budgets = budgetsResult.ok ? budgetsResult.data : [];
  const categoriesWithoutBudget = categoriesResult.ok
    ? categoriesResult.data.filter(
        (c) => c.type === "expense" && !budgets.some((b) => b.categoryId === c.id),
      )
    : [];

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="font-heading text-2xl font-semibold">Budget</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Batasi pengeluaran per kategori setiap bulan.
          </p>
        </div>
        <div className="flex gap-2">
          <CopyBudgetsButton month={month} />
          <AddBudgetDialog month={month} categories={categoriesWithoutBudget} />
        </div>
      </div>

      <nav className="flex items-center justify-between" aria-label="Pilih bulan">
        <Button asChild variant="ghost" size="icon" aria-label="Bulan sebelumnya">
          <Link href={`/budgets?month=${shiftMonth(month, -1)}`}>
            <CaretLeftIcon className="size-4" />
          </Link>
        </Button>
        <span className="font-medium">{formatMonth(month)}</span>
        <Button asChild variant="ghost" size="icon" aria-label="Bulan berikutnya">
          <Link href={`/budgets?month=${shiftMonth(month, 1)}`}>
            <CaretRightIcon className="size-4" />
          </Link>
        </Button>
      </nav>

      {!budgetsResult.ok ? (
        <p className="rounded-lg border px-4 py-16 text-center text-sm text-destructive">
          Gagal memuat budget. Silakan muat ulang halaman.
        </p>
      ) : budgets.length === 0 ? (
        <p className="rounded-lg border px-4 py-16 text-center text-sm text-muted-foreground">
          Belum ada budget untuk {formatMonth(month)}. Tambahkan budget atau salin dari bulan lalu.
        </p>
      ) : (
        <ul className="grid gap-3 sm:grid-cols-2">
          {budgets.map((b) => (
            <li key={b.id} className="rounded-lg border p-4">
              <div className="flex items-start justify-between gap-2">
                <div>
                  <p className="font-medium">{b.categoryName}</p>
                  <p className={cn("text-sm font-medium", statusColors[b.status].text)}>
                    {budgetStatusLabels[b.status]}
                  </p>
                </div>
                <BudgetRowActions budget={b} />
              </div>

              <div
                role="progressbar"
                aria-label={`Pemakaian budget ${b.categoryName}`}
                aria-valuenow={b.percentage}
                aria-valuemin={0}
                aria-valuemax={100}
                className="mt-3 h-2 overflow-hidden rounded-full bg-muted"
              >
                <div
                  className={cn("h-full rounded-full", statusColors[b.status].bar)}
                  style={{ width: `${Math.min(b.percentage, 100)}%` }}
                />
              </div>

              <p className="mt-2 text-sm tabular-nums text-muted-foreground">
                {formatRupiah(b.spent)} dari {formatRupiah(b.amount)} ({b.percentage}%)
              </p>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
