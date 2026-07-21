import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { headers } from "next/headers";
import "./globals.css";

const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] });
const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export async function generateMetadata(): Promise<Metadata> {
  const requestHeaders = await headers();
  const host =
    requestHeaders.get("x-forwarded-host") ||
    requestHeaders.get("host") ||
    "racore.xyz";
  const protocol =
    requestHeaders.get("x-forwarded-proto") ||
    (host.startsWith("localhost") ? "http" : "https");
  const image = `${protocol}://${host}/og-racore-v2.png`;
  return {
    title: "Racore — Agentic Browser & Open Web Protocol",
    description: "Browse with agents. Publish verifiable, portable websites.",
    openGraph: {
      title: "Racore.xyz",
      description: "Browse with agents. Publish without lock-in.",
      images: [
        {
          url: image,
          width: 1733,
          height: 909,
          alt: "Racore agentic browser and decentralized AI network",
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
    <html lang="en">
      <body className={`${geistSans.variable} ${geistMono.variable}`}>
        {children}
      </body>
    </html>
  );
}
