import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { AuthProvider } from "@/lib/auth-context";
import { Toaster } from "@/components/ui/sonner";

import "./globals.css";

const geistSans = Geist({
	variable: "--font-geist-sans",
	subsets: ["latin"],
});

const geistMono = Geist_Mono({
	variable: "--font-geist-mono",
	subsets: ["latin"],
});

export const metadata: Metadata = {
	title: "Vpainless",
	description: "Vpainless - vpn creation platform",
};

export default function RootLayout({
	children,
}: Readonly<{
	children: React.ReactNode;
}>) {
	return (
		<html lang="en">
			<body className={`${geistSans.variable} ${geistMono.variable} antialiased dark`} >
				<AuthProvider>{children}</AuthProvider>
				<Toaster
					position="bottom-right" // you can use: top-left, top-center, bottom-center, etc.
					richColors // optional: for prettier default styling
					closeButton // optional: adds manual close button
				/>
			</body>
		</html>
	);
}
