import type { Metadata, Viewport } from "next";
import { Plus_Jakarta_Sans } from "next/font/google";
import "./globals.css";

const jakarta = Plus_Jakarta_Sans({
  variable: "--font-jakarta",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Monetra",
  description: "Kelola keuangan pribadi dan keluarga",
};

// viewport-fit=cover lets the layout draw under the iPhone notch; the app adds safe-area padding itself.
export const viewport: Viewport = {
  viewportFit: "cover",
  themeColor: "#f7f2eb",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="id" className={`h-full antialiased ${jakarta.variable}`}>
      <body className="min-h-full flex flex-col font-sans">{children}</body>
    </html>
  );
}
