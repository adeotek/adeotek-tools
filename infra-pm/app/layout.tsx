import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "InfraPM - Infrastructure Management",
  description: "Dynamic Infrastructure Management System",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className="antialiased"
        style={{
          fontFamily: 'system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
        }}
      >
        {children}
      </body>
    </html>
  );
}
