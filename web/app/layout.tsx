import type { Metadata } from "next";
import { Providers } from "./providers";
import "@/styles/globals.css";
import AppNavbar from "@/components/layout/Navbar";
import AuthGuard from "@/components/shared/AuthGuard";

export const metadata: Metadata = {
  title: "AED QR Admin",
  description: "Admin panel for AED QR system",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="antialiased">
        <Providers>
          <AuthGuard>
            <AppNavbar />
            {children}
          </AuthGuard>
        </Providers>
      </body>
    </html>
  );
}
