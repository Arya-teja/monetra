import Link from "next/link";
import { getAccounts } from "@/features/accounts/actions";
import { getCategories } from "@/features/categories/actions";
import { cashflowTypes, type CashflowType } from "@/features/categories/types";
import { getTransactions } from "@/features/transactions/actions";
import { TransactionFormDialog } from "@/features/transactions/components/transaction-form-dialog";
import { TransactionRowActions } from "@/features/transactions/components/transaction-row-actions";
import {
  ALL,
  TransactionFilters,
} from "@/features/transactions/components/transaction-filters";
import type {
  Transaction,
  TransactionFilters as Filters,
} from "@/features/transactions/types";
import { Button } from "@/components/ui/button";
import { formatDate, formatRupiah, todayInJakarta } from "@/lib/format";
import { cn } from "@/lib/utils";

type SearchParams = Awaited<PageProps<"/transactions">["searchParams"]>;

function readParam(params: SearchParams, key: string): string | undefined {
  const value = params[key];
  const text = Array.isArray(value) ? value[0] : value;
  return text && text !== ALL ? text : undefined;
}

// Without a "month" parameter the current month is shown; an empty "month" shows all months.
function readFilters(params: SearchParams): Filters {
  const type = readParam(params, "type");
  const page = Number(readParam(params, "page"));
  return {
    month: params.month === undefined ? todayInJakarta().slice(0, 7) : readParam(params, "month"),
    accountId: readParam(params, "accountId"),
    categoryId: readParam(params, "categoryId"),
    type: cashflowTypes.includes(type as CashflowType) ? (type as CashflowType) : undefined,
    page: Number.isInteger(page) && page > 0 ? page : 1,
  };
}

function pageHref(filters: Filters, page: number) {
  const query = new URLSearchParams({ month: filters.month ?? "", page: String(page) });
  if (filters.accountId) query.set("accountId", filters.accountId);
  if (filters.categoryId) query.set("categoryId", filters.categoryId);
  if (filters.type) query.set("type", filters.type);
  return `/transactions?${query}`;
}

function groupByDate(transactions: Transaction[]) {
  const groups = new Map<string, Transaction[]>();
  for (const t of transactions) {
    groups.set(t.occurredAt, [...(groups.get(t.occurredAt) ?? []), t]);
  }
  return [...groups.entries()];
}

export default async function TransactionsPage({ searchParams }: PageProps<"/transactions">) {
  const filters = readFilters(await searchParams);
  const [transactionsResult, accountsResult, categoriesResult] = await Promise.all([
    getTransactions(filters),
    getAccounts(),
    getCategories(),
  ]);

  const accounts = accountsResult.ok ? accountsResult.data : [];
  const categories = categoriesResult.ok ? categoriesResult.data : [];
  const totalPages = transactionsResult.ok && transactionsResult.meta
    ? Math.max(1, Math.ceil(transactionsResult.meta.total / transactionsResult.meta.pageSize))
    : 1;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="font-heading text-2xl font-semibold">Transaksi</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Catat dan telusuri pemasukan serta pengeluaran kamu.
          </p>
        </div>
        <TransactionFormDialog accounts={accounts} categories={categories} />
      </div>

      <TransactionFilters filters={filters} accounts={accounts} categories={categories} />

      {!transactionsResult.ok ? (
        <p className="rounded-lg border px-4 py-16 text-center text-sm text-destructive">
          Gagal memuat transaksi. Silakan muat ulang halaman.
        </p>
      ) : transactionsResult.data.length === 0 ? (
        <p className="rounded-lg border px-4 py-16 text-center text-sm text-muted-foreground">
          Belum ada transaksi untuk filter ini.
        </p>
      ) : (
        <div className="space-y-6">
          {groupByDate(transactionsResult.data).map(([date, items]) => (
            <section key={date}>
              <h2 className="mb-2 text-sm font-medium text-muted-foreground">
                {formatDate(date)}
              </h2>
              <ul className="divide-y rounded-lg border">
                {items.map((t) => (
                  <li key={t.id} className="flex items-center gap-3 px-4 py-3">
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-medium">{t.categoryName}</p>
                      <p className="truncate text-sm text-muted-foreground">
                        {t.accountName}
                        {t.note && ` · ${t.note}`}
                      </p>
                    </div>
                    <p
                      className={cn(
                        "font-mono font-semibold tabular-nums",
                        t.type === "income" ? "text-income" : "text-expense",
                      )}
                    >
                      {t.type === "income" ? "+" : "−"}
                      {formatRupiah(t.amount)}
                    </p>
                    <TransactionRowActions
                      transaction={t}
                      accounts={accounts}
                      categories={categories}
                    />
                  </li>
                ))}
              </ul>
            </section>
          ))}

          {totalPages > 1 && (
            <nav className="flex items-center justify-between" aria-label="Halaman transaksi">
              {filters.page > 1 ? (
                <Button asChild variant="outline" className="h-11">
                  <Link href={pageHref(filters, filters.page - 1)}>Sebelumnya</Link>
                </Button>
              ) : (
                <span />
              )}
              <span className="text-sm text-muted-foreground">
                Halaman {filters.page} dari {totalPages}
              </span>
              {filters.page < totalPages ? (
                <Button asChild variant="outline" className="h-11">
                  <Link href={pageHref(filters, filters.page + 1)}>Berikutnya</Link>
                </Button>
              ) : (
                <span />
              )}
            </nav>
          )}
        </div>
      )}
    </div>
  );
}
