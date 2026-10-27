import type { Metadata } from "next";
import { headers } from "next/headers";
import "./globals.css";

export async function generateMetadata(): Promise<Metadata> {
  const requestHeaders = await headers();
  const host =
    requestHeaders.get("x-forwarded-host") ||
    requestHeaders.get("host") ||
    "racore.xyz";
  const protocol =
    requestHeaders.get("x-forwarded-proto") ||
    (host.startsWith("localhost") ? "http" : "https");
  const image = `${protocol}://${host}/og-racore-editorial.png`;
  return {
    title: "Racore — Agentic Browser & Open Web Protocol",
    description: "Browse with agents. Publish verifiable, portable websites. Keep control with Racore's local-first agentic browser and open web protocol.",
    openGraph: {
      title: "Racore.xyz",
      description: "Browse with agents. Publish without lock-in.",
      images: [
        {
          url: image,
          width: 1731,
          height: 909,
          alt: "Racore — the browser built for agency",
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title: "Racore.xyz",
      description: "Browse with agents. Publish without lock-in.",
      images: [image],
    },
  };
}

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" data-theme="piedpiper">
      <body className="racore-root">{children}</body>
    </html>
  );
}
