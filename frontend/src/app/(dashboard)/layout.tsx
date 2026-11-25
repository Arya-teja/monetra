import { redirect } from "next/navigation";
import Link from "next/link";
import { Toaster } from "sonner";
import { apiRequest } from "@/lib/api";
import { auth, signOut } from "@/lib/auth";
import { getAccounts } from "@/features/accounts/actions";
import { getCategories } from "@/features/categories/actions";
import { TransactionFormDialog } from "@/features/transactions/components/transaction-form-dialog";
import { BottomNav, SettingsLink, SidebarNav } from "@/components/layout/app-nav";
import { Button } from "@/components/ui/button";
import { PlusIcon, SignOutIcon } from "@phosphor-icons/react/ssr";

// Revokes the Go token on every device first, so a session cookie that survives
// in the browser (or is stolen) can no longer reach any data.
async function logout() {
  "use server";
  const session = await auth();
  if (session?.accessToken) {
    await apiRequest("/auth/logout", { method: "POST", token: session.accessToken });
  }
  await signOut({ redirectTo: "/login" });
}

function Logo() {
  return (
    <Link href="/dashboard" prefetch={false} className="flex items-center gap-2 font-heading text-lg font-semibold">
      <span
        aria-hidden="true"
        className="grid size-8 place-items-center rounded-lg bg-brand text-sm font-bold text-primary-foreground"
      >
        M
      </span>
      monetra
    </Link>
  );
}

export default async function DashboardLayout({ children }: { children: React.ReactNode }) {
  const [session, accountsResult, categoriesResult] = await Promise.all([
    auth(),
    getAccounts(),
    getCategories(),
  ]);
  if (!session) {
    redirect("/login");
  }

  const accounts = accountsResult.ok ? accountsResult.data : [];
  const categories = categoriesResult.ok ? categoriesResult.data : [];

  return (
    <div className="min-h-dvh bg-background lg:flex">
      <aside className="sticky top-0 hidden h-dvh w-64 shrink-0 flex-col gap-6 border-r p-4 lg:flex">
        <Logo />
        <TransactionFormDialog
          accounts={accounts}
          categories={categories}
          trigger={
            <Button className="w-full">
              <PlusIcon className="size-5" />
              Tambah Transaksi
            </Button>
          }
        />
        <SidebarNav />
        <div className="mt-auto space-y-2">
          <p className="truncate px-3 text-sm text-muted-foreground">{session.user?.name}</p>
          <form action={logout}>
            <Button type="submit" variant="ghost" className="w-full justify-start">
              <SignOutIcon className="size-5" />
              Logout
            </Button>
          </form>
        </div>
      </aside>

      <header className="sticky top-0 z-30 flex items-center justify-between border-b bg-background px-4 pt-[env(safe-area-inset-top)] lg:hidden">
        <div className="flex h-14 items-center">
          <Logo />
        </div>
        <div className="flex items-center">
          <SettingsLink />
          <form action={logout}>
            <Button type="submit" variant="ghost" size="icon" aria-label="Logout">
              <SignOutIcon className="size-6" />
            </Button>
          </form>
        </div>
      </header>

      <main className="mx-auto w-full max-w-5xl flex-1 px-4 py-6 pb-[calc(6rem+env(safe-area-inset-bottom))] lg:px-8 lg:pb-8">
        {children}
      </main>

      <BottomNav
        addButton={
          <TransactionFormDialog
            accounts={accounts}
            categories={categories}
            trigger={
              <button
                type="button"
                aria-label="Tambah transaksi"
                className="grid size-14 place-items-center rounded-full bg-primary text-primary-foreground shadow-lg outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              >
                <PlusIcon className="size-7" weight="bold" />
              </button>
            }
          />
        }
      />
      <Toaster richColors position="top-center" />
    </div>
  );
}
