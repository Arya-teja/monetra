import Link from "next/link";
import type { Budget } from "@/features/budgets/types";
import { cn } from "@/lib/utils";
import { WarningIcon } from "@phosphor-icons/react/ssr";

// In-app warning for budgets that are 80% used or more (PRD F5).
export function BudgetAlertBanner({ alerts }: { alerts: Budget[] }) {
  if (alerts.length === 0) return null;

  const anyOver = alerts.some((b) => b.status === "over");

  return (
    <div
      role="status"
      className={cn(
        "flex gap-3 rounded-lg border p-4",
        anyOver ? "border-expense bg-expense/10" : "border-warning bg-warning/10",
      )}
    >
      <WarningIcon
        aria-hidden="true"
        className={cn("mt-0.5 size-5 shrink-0", anyOver ? "text-expense" : "text-warning")}
      />
      <div className="text-sm">
        <p className="font-medium">Perhatikan budget bulan ini</p>
        <ul className="mt-1 space-y-0.5">
          {alerts.map((b) => (
            <li key={b.id}>
              {b.categoryName}:{" "}
              {b.status === "over" ? `terlampaui (${b.percentage}%)` : `terpakai ${b.percentage}%`}
            </li>
          ))}
        </ul>
        <Link href="/budgets" className="mt-2 inline-block font-medium underline">
          Lihat budget
        </Link>
      </div>
    </div>
  );
}
