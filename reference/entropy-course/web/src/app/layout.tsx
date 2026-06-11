import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Entropy Course Reader",
  description: "Learner-facing course reader for published Entropy courses.",
  icons: {
    icon: "/icon.svg",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <a className="skip-link" href="#main-content">
          Skip to Content
        </a>
        {children}
      </body>
    </html>
  );
}
