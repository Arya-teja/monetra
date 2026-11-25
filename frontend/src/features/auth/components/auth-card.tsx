import Link from "next/link";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

type Props = {
  description: string;
  footer: { text: string; linkLabel: string; href: string };
  children: React.ReactNode;
};

// Card around the login and register forms: same heading, page-specific text and link.
export function AuthCard({ description, footer, children }: Props) {
  return (
    <Card className="w-full overflow-visible border-0 bg-transparent py-0 ring-0">
      <CardHeader className="px-0 pb-7">
        <p className="mb-3 text-xs font-semibold uppercase tracking-[0.18em] text-primary">
          Start with clarity
        </p>
        <CardTitle className="font-heading text-3xl font-semibold tracking-tight sm:text-4xl">
          Buat ruang untuk tujuanmu.
        </CardTitle>
        <CardDescription className="mt-2 max-w-sm text-sm leading-6">{description}</CardDescription>
      </CardHeader>
      <CardContent className="px-0">{children}</CardContent>
      <CardFooter className="mt-7 justify-center gap-x-0.5 border-t px-0 pt-6 text-sm text-muted-foreground">
        {footer.text}{" "}
        <Link href={footer.href} className="font-semibold text-primary hover:underline">
          {footer.linkLabel}
        </Link>
      </CardFooter>
    </Card>
  );
}
