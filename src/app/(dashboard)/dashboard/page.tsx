import { auth } from "@/lib/auth";

export default async function DashboardPage() {
  const session = await auth();

  return (
    <div>
      <h1 className="font-heading text-2xl font-semibold">
        Halo, {session?.user?.name}
      </h1>
      <p className="mt-2 text-muted-foreground">
        Ini halaman dashboard Monetra — akan diisi ringkasan saldo, chart, dan
        transaksi terbaru.
      </p>
    </div>
  );
}
