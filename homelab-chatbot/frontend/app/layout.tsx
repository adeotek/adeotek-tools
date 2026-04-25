import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "homelab-chatbot",
  description: "Home-lab knowledge base chatbot",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-neutral-950 text-neutral-100 antialiased">{children}</body>
    </html>
  );
}
