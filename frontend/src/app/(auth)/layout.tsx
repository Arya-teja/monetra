// Shared frame of /login and /register: brand panel on the left, the page's card on the right.
export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="min-h-screen bg-background px-4 py-5 text-foreground sm:px-6 lg:px-10 lg:py-8">
      <div className="mx-auto grid min-h-[calc(100vh-2.5rem)] max-w-6xl overflow-hidden rounded-2xl border bg-background shadow-lg lg:grid-cols-[1.05fr_0.95fr]">
        <section className="relative flex flex-col justify-between overflow-hidden bg-primary p-7 text-primary-foreground lg:min-h-105 sm:p-10 lg:p-12">
          <div className="absolute -right-24 -top-24 size-72 rounded-full border border-primary-foreground/20" />
          <div className="absolute -right-8 top-8 size-40 rounded-full border border-primary-foreground/15" />

          <div className="relative z-10">
            <div className="flex items-center gap-3">
              <div className="grid size-9 place-items-center rounded-lg bg-background text-sm font-black text-primary">
                M
              </div>
              <span className="font-heading text-lg font-semibold tracking-tight">monetra</span>
            </div>
            <div className="mt-8 max-w-md lg:mt-24">
              <p className="mb-4 text-xs font-semibold uppercase tracking-[0.22em] text-primary-foreground">
                Your money, in focus
              </p>
              <h1 className="font-heading text-3xl font-semibold leading-[1.05] tracking-tight lg:text-5xl">
                Make every rupiah tell a better story.
              </h1>
              <p className="mt-5 max-w-sm text-sm leading-6 text-primary-foreground/90">
                Satu tempat untuk memahami kebiasaan belanja, merencanakan tujuan, dan menjaga uang
                tetap bergerak ke arah yang benar.
              </p>
            </div>
          </div>
        </section>

        <section className="flex items-center bg-background px-6 py-10 sm:px-12 lg:px-16">
          {children}
        </section>
      </div>
    </main>
  );
}
