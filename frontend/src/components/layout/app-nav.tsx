"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  ChartPieSliceIcon,
  GearIcon,
  HouseIcon,
  ReceiptIcon,
  WalletIcon,
  type Icon,
} from "@phosphor-icons/react";
import { cn } from "@/lib/utils";

type NavItem = { href: string; label: string; icon: Icon };

const HOME: NavItem = { href: "/dashboard", label: "Beranda", icon: HouseIcon };
const TRANSACTIONS: NavItem = { href: "/transactions", label: "Transaksi", icon: ReceiptIcon };
const BUDGETS: NavItem = { href: "/budgets", label: "Budget", icon: ChartPieSliceIcon };
const ACCOUNTS: NavItem = { href: "/accounts", label: "Akun", icon: WalletIcon };
const SETTINGS: NavItem = { href: "/settings", label: "Pengaturan", icon: GearIcon };

function NavLink({ item, layout }: { item: NavItem; layout: "bottom" | "sidebar" }) {
  const active = usePathname().startsWith(item.href);
  const Icon = item.icon;

  return (
    <Link
      href={item.href}
      // No prefetch: prefetch responses that arrive after a logout would
      // re-set the session cookie (each response refreshes it).
      prefetch={false}
      aria-current={active ? "page" : undefined}
      className={cn(
        "flex items-center rounded-lg font-medium transition-colors",
        layout === "bottom"
          ? "min-h-14 flex-col justify-center gap-0.5 text-xs"
          : "h-11 gap-3 px-3 text-sm hover:bg-accent",
        active ? "text-primary" : "text-muted-foreground",
        active && layout === "sidebar" && "bg-accent",
      )}
    >
      <Icon className="size-6" weight={active ? "fill" : "regular"} aria-hidden="true" />
      {item.label}
    </Link>
  );
}

// Phone navigation (PRD §9): Beranda · Transaksi · (+) · Budget · Akun.
export function BottomNav({ addButton }: { addButton: React.ReactNode }) {
  return (
    <nav
      aria-label="Navigasi utama"
      className="fixed inset-x-0 bottom-0 z-40 border-t bg-background pb-[env(safe-area-inset-bottom)] lg:hidden"
    >
      <div className="mx-auto grid max-w-md grid-cols-5 items-center px-2">
        <NavLink item={HOME} layout="bottom" />
        <NavLink item={TRANSACTIONS} layout="bottom" />
        <div className="flex justify-center">{addButton}</div>
        <NavLink item={BUDGETS} layout="bottom" />
        <NavLink item={ACCOUNTS} layout="bottom" />
      </div>
    </nav>
  );
}

export function SidebarNav() {
  return (
    <nav aria-label="Navigasi utama" className="flex flex-col gap-1">
      {[HOME, TRANSACTIONS, BUDGETS, ACCOUNTS, SETTINGS].map((item) => (
        <NavLink key={item.href} item={item} layout="sidebar" />
      ))}
    </nav>
  );
}

// On phones, Pengaturan is not in the bottom bar, so the top bar links to it.
export function SettingsLink() {
  const active = usePathname().startsWith(SETTINGS.href);
  return (
    <Link
      href={SETTINGS.href}
      prefetch={false}
      aria-label="Pengaturan"
      aria-current={active ? "page" : undefined}
      className={cn(
        "flex size-11 items-center justify-center rounded-lg hover:bg-accent",
        active ? "text-primary" : "text-muted-foreground",
      )}
    >
      <GearIcon className="size-6" aria-hidden="true" />
    </Link>
  );
}
