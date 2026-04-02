"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/contexts/auth-context";
import {
  getSellerDashboard,
  getBuyerDashboard,
  type SellerDashboard,
  type BuyerDashboard,
} from "@/lib/api";

function StatCard({
  label,
  value,
  sub,
  color = "emerald",
}: {
  label: string;
  value: string | number;
  sub?: string;
  color?: "emerald" | "blue" | "purple" | "amber";
}) {
  const colorMap = {
    emerald: "text-emerald-400",
    blue: "text-blue-400",
    purple: "text-purple-400",
    amber: "text-amber-400",
  };

  return (
    <div className="bg-white/5 border border-white/10 rounded-xl p-5">
      <p className="text-xs text-zinc-500 uppercase tracking-wider mb-1">
        {label}
      </p>
      <p className={`text-2xl font-bold ${colorMap[color]}`}>{value}</p>
      {sub && <p className="text-xs text-zinc-500 mt-1">{sub}</p>}
    </div>
  );
}

function SellerView() {
  const [data, setData] = useState<SellerDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getSellerDashboard()
      .then(setData)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <LoadingState />;
  }
  if (error) {
    return <ErrorState message={error} />;
  }
  if (!data) return null;

  const epsilonPct =
    data.epsilon_budget > 0
      ? ((data.epsilon_spent / data.epsilon_budget) * 100).toFixed(1)
      : "0";

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold mb-1">Seller Dashboard</h1>
        <p className="text-zinc-500">
          Welcome back, {data.user.display_name}
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          label="Credit Balance"
          value={data.credit_balance.toLocaleString()}
          color="emerald"
        />
        <StatCard
          label="Epsilon Budget"
          value={`${data.epsilon_remaining.toFixed(2)} / ${data.epsilon_budget.toFixed(1)}`}
          sub={`${epsilonPct}% used`}
          color="blue"
        />
        <StatCard
          label="Total Batches"
          value={data.total_batches}
          color="purple"
        />
        <StatCard
          label="Recent Earnings"
          value={
            data.recent_transactions
              .filter((t) => t.amount > 0)
              .reduce((sum, t) => sum + t.amount, 0)
              .toLocaleString()
          }
          sub="From recent transactions"
          color="amber"
        />
      </div>

      {/* Epsilon gauge */}
      <div className="bg-white/5 border border-white/10 rounded-xl p-5">
        <h2 className="text-sm font-semibold text-zinc-300 mb-3">
          Privacy Budget
        </h2>
        <div className="w-full bg-white/10 rounded-full h-3">
          <div
            className="bg-blue-500 h-3 rounded-full transition-all duration-500"
            style={{ width: `${Math.min(Number(epsilonPct), 100)}%` }}
          />
        </div>
        <div className="flex justify-between mt-2 text-xs text-zinc-500">
          <span>{data.epsilon_spent.toFixed(3)} spent</span>
          <span>{data.epsilon_remaining.toFixed(3)} remaining</span>
        </div>
      </div>

      {/* Recent batches */}
      <div className="bg-white/5 border border-white/10 rounded-xl p-5">
        <h2 className="text-sm font-semibold text-zinc-300 mb-4">
          Recent Batches
        </h2>
        {data.recent_batches && data.recent_batches.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-zinc-500 border-b border-white/10">
                  <th className="text-left py-2 font-medium">ID</th>
                  <th className="text-left py-2 font-medium">Records</th>
                  <th className="text-left py-2 font-medium">Status</th>
                  <th className="text-left py-2 font-medium">Created</th>
                </tr>
              </thead>
              <tbody>
                {data.recent_batches.map((batch) => (
                  <tr
                    key={batch.id}
                    className="border-b border-white/5 last:border-0"
                  >
                    <td className="py-2.5 font-mono text-xs text-zinc-400">
                      {batch.id.slice(0, 8)}...
                    </td>
                    <td className="py-2.5">{batch.record_count}</td>
                    <td className="py-2.5">
                      <StatusBadge status={batch.status} />
                    </td>
                    <td className="py-2.5 text-zinc-500">
                      {new Date(batch.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-zinc-500 text-sm">No batches uploaded yet.</p>
        )}
      </div>

      {/* Recent transactions */}
      <div className="bg-white/5 border border-white/10 rounded-xl p-5">
        <h2 className="text-sm font-semibold text-zinc-300 mb-4">
          Recent Transactions
        </h2>
        {data.recent_transactions && data.recent_transactions.length > 0 ? (
          <div className="space-y-2">
            {data.recent_transactions.map((tx) => (
              <div
                key={tx.id}
                className="flex items-center justify-between py-2 border-b border-white/5 last:border-0"
              >
                <div>
                  <p className="text-sm">{tx.description || tx.tx_type}</p>
                  <p className="text-xs text-zinc-500">
                    {new Date(tx.created_at).toLocaleString()}
                  </p>
                </div>
                <span
                  className={`font-mono font-medium ${
                    tx.amount > 0 ? "text-emerald-400" : "text-red-400"
                  }`}
                >
                  {tx.amount > 0 ? "+" : ""}
                  {tx.amount.toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-zinc-500 text-sm">No transactions yet.</p>
        )}
      </div>
    </div>
  );
}

function BuyerView() {
  const [data, setData] = useState<BuyerDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getBuyerDashboard()
      .then(setData)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <LoadingState />;
  }
  if (error) {
    return <ErrorState message={error} />;
  }
  if (!data) return null;

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold mb-1">Buyer Dashboard</h1>
        <p className="text-zinc-500">
          Welcome back, {data.user.display_name}
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard
          label="Credit Balance"
          value={data.credit_balance.toLocaleString()}
          color="emerald"
        />
        <StatCard
          label="Total Purchases"
          value={data.total_purchases}
          color="blue"
        />
        <StatCard
          label="Active Bids"
          value={data.active_bids}
          color="purple"
        />
      </div>

      {/* Recent purchases */}
      <div className="bg-white/5 border border-white/10 rounded-xl p-5">
        <h2 className="text-sm font-semibold text-zinc-300 mb-4">
          Recent Purchases
        </h2>
        {data.recent_purchases && data.recent_purchases.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-zinc-500 border-b border-white/10">
                  <th className="text-left py-2 font-medium">Dataset</th>
                  <th className="text-left py-2 font-medium">Price</th>
                  <th className="text-left py-2 font-medium">Status</th>
                  <th className="text-left py-2 font-medium">Date</th>
                </tr>
              </thead>
              <tbody>
                {data.recent_purchases.map((p) => (
                  <tr
                    key={p.id}
                    className="border-b border-white/5 last:border-0"
                  >
                    <td className="py-2.5 font-mono text-xs text-zinc-400">
                      {p.dataset_id.slice(0, 8)}...
                    </td>
                    <td className="py-2.5">{p.price_credits} credits</td>
                    <td className="py-2.5">
                      <StatusBadge status={p.status} />
                    </td>
                    <td className="py-2.5 text-zinc-500">
                      {new Date(p.purchased_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-zinc-500 text-sm">No purchases yet.</p>
        )}
      </div>

      {/* Recent transactions */}
      <div className="bg-white/5 border border-white/10 rounded-xl p-5">
        <h2 className="text-sm font-semibold text-zinc-300 mb-4">
          Recent Transactions
        </h2>
        {data.recent_transactions && data.recent_transactions.length > 0 ? (
          <div className="space-y-2">
            {data.recent_transactions.map((tx) => (
              <div
                key={tx.id}
                className="flex items-center justify-between py-2 border-b border-white/5 last:border-0"
              >
                <div>
                  <p className="text-sm">{tx.description || tx.tx_type}</p>
                  <p className="text-xs text-zinc-500">
                    {new Date(tx.created_at).toLocaleString()}
                  </p>
                </div>
                <span
                  className={`font-mono font-medium ${
                    tx.amount > 0 ? "text-emerald-400" : "text-red-400"
                  }`}
                >
                  {tx.amount > 0 ? "+" : ""}
                  {tx.amount.toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-zinc-500 text-sm">No transactions yet.</p>
        )}
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    raw: "bg-zinc-500/20 text-zinc-400",
    validated: "bg-blue-500/20 text-blue-400",
    anonymized: "bg-purple-500/20 text-purple-400",
    listed: "bg-emerald-500/20 text-emerald-400",
    sold: "bg-amber-500/20 text-amber-400",
    withdrawn: "bg-red-500/20 text-red-400",
    completed: "bg-emerald-500/20 text-emerald-400",
    pending: "bg-amber-500/20 text-amber-400",
    active: "bg-blue-500/20 text-blue-400",
    failed: "bg-red-500/20 text-red-400",
  };

  return (
    <span
      className={`inline-block text-xs px-2 py-0.5 rounded-full font-medium ${
        colorMap[status] || "bg-zinc-500/20 text-zinc-400"
      }`}
    >
      {status}
    </span>
  );
}

function LoadingState() {
  return (
    <div className="flex items-center justify-center py-20">
      <div className="w-8 h-8 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
    </div>
  );
}

function ErrorState({ message }: { message: string }) {
  return (
    <div className="bg-red-500/10 border border-red-500/20 text-red-400 rounded-xl p-6 text-center">
      <p className="font-medium mb-1">Something went wrong</p>
      <p className="text-sm">{message}</p>
    </div>
  );
}

export default function DashboardPage() {
  const { user } = useAuth();

  if (!user) return null;

  return user.role === "seller" ? <SellerView /> : <BuyerView />;
}
