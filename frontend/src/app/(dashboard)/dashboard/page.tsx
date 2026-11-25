import Link from "next/link";
import { auth } from "@/lib/auth";
import { getDashboardSummary, getTrend } from "@/features/dashboard/actions";
import { BudgetAlertBanner } from "@/features/dashboard/components/budget-alert-banner";
import { TrendChart } from "@/features/dashboard/components/trend-chart";
import { formatDate, formatMonth, formatRupiah } from "@/lib/format";
import { cn } from "@/lib/utils";

function StatCard({ label, value, className }: { label: string; value: string; className?: string }) {
  return (
    <div className="rounded-lg border p-4">
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className={cn("mt-1 text-xl font-semibold tabular-nums", className)}>{value}</p>
    </div>
  );
}

export default async function DashboardPage() {
  const [session, summaryResult, trendResult] = await Promise.all([
    auth(),
    getDashboardSummary(),
    getTrend(),
  ]);

  if (!summaryResult.ok) {
    return (
      <p className="rounded-lg border px-4 py-16 text-center text-sm text-destructive">
        Gagal memuat dashboard. Silakan muat ulang halaman.
      </p>
    );
  }
  const s = summaryResult.data;
  const topMax = Math.max(1, ...s.topCategories.map((c) => c.total));

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-heading text-2xl font-semibold">Halo, {session?.user?.name}</h1>
        <p className="mt-1 text-sm text-muted-foreground">Ringkasan {formatMonth(s.month)}</p>
      </div>

      <BudgetAlertBanner alerts={s.budgetAlerts} />

      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <StatCard
          label="Total saldo"
          value={formatRupiah(s.totalBalance)}
          className={s.totalBalance < 0 ? "text-expense" : undefined}
        />
        <StatCard label="Pemasukan" value={formatRupiah(s.income)} className="text-income" />
        <StatCard label="Pengeluaran" value={formatRupiah(s.expense)} className="text-expense" />
        <StatCard
          label="Income terpakai"
          value={s.usedPercentage === null ? "—" : `${s.usedPercentage}%`}
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <section className="rounded-lg border p-4">
          <h2 className="mb-4 font-medium">Pemasukan vs Pengeluaran (6 bulan)</h2>
          {trendResult.ok ? (
            <TrendChart months={trendResult.data} />
          ) : (
            <p className="text-sm text-destructive">Gagal memuat grafik.</p>
          )}
        </section>

        <section className="rounded-lg border p-4">
          <h2 className="mb-4 font-medium">Top 5 pengeluaran bulan ini</h2>
          {s.topCategories.length === 0 ? (
            <p className="text-sm text-muted-foreground">Belum ada pengeluaran bulan ini.</p>
          ) : (
            <ul className="space-y-3">
              {s.topCategories.map((c) => (
                <li key={c.categoryId}>
                  <div className="flex justify-between gap-2 text-sm">
                    <span>{c.categoryName}</span>
                    <span className="tabular-nums">{formatRupiah(c.total)}</span>
                  </div>
                  <div aria-hidden="true" className="mt-1 h-1.5 overflow-hidden rounded-full bg-muted">
                    <div
                      className="h-full rounded-full bg-expense"
                      style={{ width: `${(c.total / topMax) * 100}%` }}
                    />
                  </div>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>

      <section className="rounded-lg border">
        <div className="flex items-center justify-between px-4 pt-4">
          <h2 className="font-medium">Transaksi terakhir</h2>
          <Link href="/transactions" className="text-sm font-medium text-primary hover:underline">
            Lihat semua
          </Link>
        </div>
        {s.recentTransactions.length === 0 ? (
          <p className="px-4 py-8 text-sm text-muted-foreground">Belum ada transaksi.</p>
        ) : (
          <ul className="mt-2 divide-y">
            {s.recentTransactions.map((t) => (
              <li key={t.id} className="flex items-center gap-3 px-4 py-3">
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium">{t.categoryName}</p>
                  <p className="truncate text-sm text-muted-foreground">
                    {formatDate(t.occurredAt)} · {t.accountName}
                  </p>
                </div>
                <p
                  className={cn(
                    "font-semibold tabular-nums",
                    t.type === "income" ? "text-income" : "text-expense",
                  )}
                >
                  {t.type === "income" ? "+" : "−"}
                  {formatRupiah(t.amount)}
                </p>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
