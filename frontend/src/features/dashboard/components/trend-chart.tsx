import type { MonthTotal } from "../types";
import { formatMonth, formatMonthShort, formatRupiah } from "@/lib/format";

// Bar chart of income vs expense per month, built from plain elements (no chart library).
// The bars are decorative; screen readers get the same numbers from a hidden table.
export function TrendChart({ months }: { months: MonthTotal[] }) {
  const max = Math.max(1, ...months.flatMap((m) => [m.income, m.expense]));
  const height = (value: number) => `${(value / max) * 100}%`;

  return (
    <div>
      <div aria-hidden="true" className="flex h-40 items-end gap-2">
        {months.map((m) => (
          <div key={m.month} className="flex h-full flex-1 flex-col">
            <div className="flex flex-1 items-end justify-center gap-1">
              <div className="w-3 rounded-t bg-income" style={{ height: height(m.income) }} />
              <div className="w-3 rounded-t bg-expense" style={{ height: height(m.expense) }} />
            </div>
            <p className="mt-1 text-center text-xs text-muted-foreground">
              {formatMonthShort(m.month)}
            </p>
          </div>
        ))}
      </div>

      <div aria-hidden="true" className="mt-3 flex gap-4 text-xs text-muted-foreground">
        <span className="flex items-center gap-1">
          <span className="size-2.5 rounded-sm bg-income" /> Pemasukan
        </span>
        <span className="flex items-center gap-1">
          <span className="size-2.5 rounded-sm bg-expense" /> Pengeluaran
        </span>
      </div>

      <table className="sr-only">
        <caption>Pemasukan dan pengeluaran per bulan</caption>
        <thead>
          <tr>
            <th scope="col">Bulan</th>
            <th scope="col">Pemasukan</th>
            <th scope="col">Pengeluaran</th>
          </tr>
        </thead>
        <tbody>
          {months.map((m) => (
            <tr key={m.month}>
              <th scope="row">{formatMonth(m.month)}</th>
              <td>{formatRupiah(m.income)}</td>
              <td>{formatRupiah(m.expense)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
