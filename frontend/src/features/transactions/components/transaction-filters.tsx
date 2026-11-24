import Link from "next/link";
import type { TransactionFilters as Filters } from "../types";
import type { Account } from "@/features/accounts/types";
import {
  cashflowTypeLabels,
  cashflowTypes,
  type Category,
} from "@/features/categories/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

// "all" is sent for an empty select, because a select option cannot have an empty value.
export const ALL = "all";

type Props = {
  filters: Filters;
  accounts: Account[];
  categories: Category[];
};

// A plain GET form: applying filters updates the URL, so filtered views can be bookmarked.
export function TransactionFilters({ filters, accounts, categories }: Props) {
  return (
    <form
      method="get"
      action="/transactions"
      className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5 lg:items-end"
    >
      <div className="space-y-2">
        <Label htmlFor="filter-month">Bulan</Label>
        <Input
          id="filter-month"
          name="month"
          type="month"
          defaultValue={filters.month ?? ""}
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="filter-type">Tipe</Label>
        <Select name="type" defaultValue={filters.type ?? ALL}>
          <SelectTrigger id="filter-type" className="w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>Semua tipe</SelectItem>
            {cashflowTypes.map((t) => (
              <SelectItem key={t} value={t}>
                {cashflowTypeLabels[t]}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="filter-account">Akun</Label>
        <Select name="accountId" defaultValue={filters.accountId ?? ALL}>
          <SelectTrigger id="filter-account" className="w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>Semua akun</SelectItem>
            {accounts.map((a) => (
              <SelectItem key={a.id} value={a.id}>
                {a.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="filter-category">Kategori</Label>
        <Select name="categoryId" defaultValue={filters.categoryId ?? ALL}>
          <SelectTrigger id="filter-category" className="w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>Semua kategori</SelectItem>
            {categories.map((c) => (
              <SelectItem key={c.id} value={c.id}>
                {c.name} ({cashflowTypeLabels[c.type]})
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="flex gap-2">
        <Button type="submit" className="h-11 flex-1">
          Terapkan
        </Button>
        <Button asChild variant="outline" className="h-11">
          <Link href="/transactions">Reset</Link>
        </Button>
      </div>
    </form>
  );
}
