"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";
import { AuthProvider, useAuth } from "@/contexts/auth-context";
import { SolanaProvider } from "@/contexts/solana-context";
import { WalletButton } from "@/components/solana/WalletButton";

const SELLER_NAV = [
  { label: "Dashboard", href: "/dashboard", icon: "M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0a1 1 0 01-1-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 01-1 1" },
  { label: "My Data", href: "/my-data", icon: "M4 7v10c0 2 1 3 3 3h10c2 0 3-1 3-3V7M4 7l4-4h8l4 4M9 11h6" },
  { label: "Privacy", href: "/privacy", icon: "M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" },
  { label: "Marketplace", href: "/marketplace", icon: "M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" },
  { label: "Solana", href: "/solana", icon: "M13 10V3L4 14h7v7l9-11h-7" },
];

const BUYER_NAV = [
  { label: "Dashboard", href: "/dashboard", icon: "M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0a1 1 0 01-1-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 01-1 1" },
  { label: "Marketplace", href: "/marketplace", icon: "M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" },
  { label: "My Purchases", href: "/my-data", icon: "M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" },
  { label: "My Bids", href: "/privacy", icon: "M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" },
  { label: "Solana", href: "/solana", icon: "M13 10V3L4 14h7v7l9-11h-7" },
];

function DashboardShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, loading, logout } = useAuth();

  useEffect(() => {
    if (!loading && !user) {
      router.push("/login");
    }
  }, [loading, user, router]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (!user) return null;

  const nav = user.role === "seller" ? SELLER_NAV : BUYER_NAV;

  return (
    <div className="min-h-screen flex">
      {/* Sidebar */}
      <aside className="w-64 bg-white/[0.02] border-r border-white/10 flex flex-col">
        <div className="px-6 py-5 border-b border-white/10">
          <Link href="/" className="text-xl font-bold tracking-tight">
            Screener
          </Link>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-1">
          {nav.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-white/10 text-white"
                    : "text-zinc-400 hover:text-white hover:bg-white/5"
                }`}
              >
                <svg
                  className="w-5 h-5 shrink-0"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  strokeWidth={1.5}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d={item.icon}
                  />
                </svg>
                {item.label}
              </Link>
            );
          })}
        </nav>

        {/* Wallet & Credits */}
        <div className="px-4 py-4 border-t border-white/10 space-y-3">
          <WalletButton />

          <div className="bg-white/5 rounded-lg px-4 py-3">
            <p className="text-xs text-zinc-500 uppercase tracking-wider mb-1">
              Credits
            </p>
            <p className="text-lg font-bold text-emerald-400">
              {user.credit_balance.toLocaleString()}
            </p>
          </div>

          <div className="flex items-center justify-between">
            <div className="min-w-0">
              <p className="text-sm font-medium truncate">
                {user.display_name}
              </p>
              <p className="text-xs text-zinc-500 truncate">{user.email}</p>
            </div>
            <button
              onClick={logout}
              className="text-zinc-500 hover:text-red-400 transition-colors shrink-0 ml-2"
              title="Log out"
            >
              <svg
                className="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                strokeWidth={1.5}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15m3-3h-9m9 0l-3-3m3 3l-3 3"
                />
              </svg>
            </button>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto">
        <div className="max-w-5xl mx-auto px-8 py-8">{children}</div>
      </main>
    </div>
  );
}

export default function DashboardLayout({
  children,
}: {
  children: ReactNode;
}) {
  return (
    <AuthProvider>
      <SolanaProvider>
        <DashboardShell>{children}</DashboardShell>
      </SolanaProvider>
    </AuthProvider>
  );
}
