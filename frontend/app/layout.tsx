import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Chia Bill — Split the Bill",
  description: "Itemize the table, tap who ordered what, print the split.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
