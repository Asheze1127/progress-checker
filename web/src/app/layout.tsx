import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Progress Checker",
  description: "Hackathon progress tracking dashboard",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body>{children}</body>
    </html>
  );
}
