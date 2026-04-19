import type { NextConfig } from "next";

const isDev = process.env.NODE_ENV === "development";

const nextConfig: NextConfig = {
  // Static export for production (served by FastAPI StaticFiles).
  // Omitted in dev so Next.js can proxy /api/* to the backend.
  ...(!isDev && { output: "export" }),
  images: { unoptimized: true },
  trailingSlash: true,
  async rewrites() {
    if (!isDev) return [];
    return [{ source: "/api/:path*", destination: "http://localhost:8000/api/:path*" }];
  },
};

export default nextConfig;
