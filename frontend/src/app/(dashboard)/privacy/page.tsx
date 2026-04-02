"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/contexts/auth-context";
import {
  getPrivacyBudget,
  getEpsilonLedger,
  getBids,
  type PrivacyBudget,
  type EpsilonLedgerEntry,
  type Bid,
} from "@/lib/api";

// ---------------------------------------------------------------------------
// Seller: Privacy Budget Dashboard
// ---------------------------------------------------------------------------

function SellerPrivacyPage() {
  const [budget, setBudget] = useState<PrivacyBudget | null>(null);
  const [entries, setEntries] = useState<EpsilonLedgerEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const [budgetData, ledgerData] = await Promise.all([
          getPrivacyBudget(),
          getEpsilonLedger(50, 0),
        ]);
        setBudget(budgetData);
        setEntries(ledgerData.entries);
        setTotal(ledgerData.total);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load privacy data");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="w-8 h-8 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-500/10 border border-red-500/20 text-red-400 rounded-xl p-6 text-center">
        <p className="font-medium mb-1">Something went wrong</p>
        <p className="text-sm">{error}</p>
      </div>
    );
  }

  const pctUsed = budget?.percent_used ?? 0;

  // Determine color based on usage
  let barColor = "bg-emerald-500";
  let textColor = "text-emerald-400";
  if (pctUsed > 75) {
    barColor = "bg-red-500";
    textColor = "text-red-400";
  } else if (pctUsed > 50) {
    barColor = "bg-amber-500";
    textColor = "text-amber-400";
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold mb-1">Privacy Dashboard</h1>
        <p className="text-zinc-500">
          Monitor your differential privacy epsilon budget
        </p>
      </div>

      {/* Budget overview cards */}
      {budget && (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div className="bg-white/5 border border-white/10 rounded-xl p-5">
            <p className="text-xs text-zinc-500 uppercase tracking-wider mb-1">
              Total Budget
            </p>
            <p className="text-2xl font-bold text-blue-400">
              {budget.total_budget.toFixed(2)}
            </p>
          </div>
          <div className="bg-white/5 border border-white/10 rounded-xl p-5">
            <p className="text-xs text-zinc-500 uppercase tracking-wider mb-1">
              Epsilon Spent
            </p>
            <p className={`text-2xl font-bold ${textColor}`}>
              {budget.epsilon_spent.toFixed(4)}
            </p>
          </div>
          <div className="bg-white/5 border border-white/10 rounded-xl p-5">
            <p className="text-xs text-zinc-500 uppercase tracking-wider mb-1">
              Remaining
            </p>
            <p className="text-2xl font-bold text-emerald-400">
              {budget.epsilon_remaining.toFixed(4)}
            </p>
          </div>
        </div>
      )}

      {/* Epsilon budget gauge */}
      {budget && (
        <div className="bg-white/5 border border-white/10 rounded-xl p-6">
          <h2 className="text-sm font-semibold text-zinc-300 mb-4">
            Budget Usage
          </h2>
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="text-zinc-400">
                {budget.epsilon_spent.toFixed(4)} / {budget.total_budget.toFixed(2)}
              </span>
              <span className={`font-medium ${textColor}`}>
                {pctUsed.toFixed(1)}% used
              </span>
            </div>

            {/* Progress bar */}
            <div className="w-full bg-white/10 rounded-full h-4 overflow-hidden">
              <div
                className={`${barColor} h-4 rounded-full transition-all duration-700 ease-out`}
                style={{ width: `${Math.min(pctUsed, 100)}%` }}
              />
            </div>

            {/* Scale markers */}
            <div className="flex justify-between text-xs text-zinc-600">
              <span>0</span>
              <span>25%</span>
              <span>50%</span>
              <span>75%</span>
              <span>100%</span>
            </div>
          </div>

          {pctUsed > 75 && (
            <div className="mt-4 bg-red-500/10 border border-red-500/20 rounded-lg p-3 text-sm text-red-400">
              Your privacy budget is running low. Once exhausted, your data
              cannot be included in new datasets.
            </div>
          )}
        </div>
      )}

      {/* Epsilon ledger */}
      <div className="bg-white/5 border border-white/10 rounded-xl p-6">
        <h2 className="text-sm font-semibold text-zinc-300 mb-4">
          Epsilon Ledger ({total} entries)
        </h2>

        {entries.length === 0 ? (
          <p className="text-zinc-500 text-sm">No epsilon events recorded yet.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-zinc-500 border-b border-white/10">
                  <th className="text-left py-2 font-medium">Event</th>
                  <th className="text-left py-2 font-medium">Epsilon</th>
                  <th className="text-left py-2 font-medium">Remaining</th>
                  <th className="text-left py-2 font-medium">Dataset</th>
                  <th className="text-left py-2 font-medium">Description</th>
                  <th className="text-left py-2 font-medium">Date</th>
                </tr>
              </thead>
              <tbody>
                {entries.map((entry) => (
                  <tr
                    key={entry.id}
                    className="border-b border-white/5 last:border-0"
                  >
                    <td className="py-3">
                      <EventBadge type={entry.event_type} />
                    </td>
                    <td className="py-3 font-mono text-red-400">
                      -{entry.epsilon_spent.toFixed(4)}
                    </td>
                    <td className="py-3 font-mono text-zinc-400">
                      {entry.epsilon_remaining.toFixed(4)}
                    </td>
                    <td className="py-3 font-mono text-xs text-zinc-500">
                      {entry.dataset_id
                        ? `${entry.dataset_id.slice(0, 8)}...`
                        : "--"}
                    </td>
                    <td className="py-3 text-zinc-400 text-xs max-w-48 truncate">
                      {entry.description || "--"}
                    </td>
                    <td className="py-3 text-zinc-500">
                      {new Date(entry.created_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

function EventBadge({ type }: { type: string }) {
  const colorMap: Record<string, string> = {
    dataset_sale: "bg-emerald-500/20 text-emerald-400",
    query_response: "bg-blue-500/20 text-blue-400",
    sample_generation: "bg-purple-500/20 text-purple-400",
    budget_refund: "bg-amber-500/20 text-amber-400",
  };

  const labelMap: Record<string, string> = {
    dataset_sale: "Sale",
    query_response: "Query",
    sample_generation: "Sample",
    budget_refund: "Refund",
  };

  return (
    <span
      className={`inline-block text-xs px-2 py-0.5 rounded-full font-medium ${
        colorMap[type] || "bg-zinc-500/20 text-zinc-400"
      }`}
    >
      {labelMap[type] || type}
    </span>
  );
}

// ---------------------------------------------------------------------------
// Buyer: My Bids
// ---------------------------------------------------------------------------

function BuyerBidsPage() {
  const [bids, setBids] = useState<Bid[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getBids()
      .then(setBids)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold mb-1">My Bids</h1>
        <p className="text-zinc-500">
          Track your bids on custom data segments
        </p>
      </div>

      <div className="bg-white/5 border border-white/10 rounded-xl p-6">
        {loading ? (
          <div className="flex items-center justify-center py-10">
            <div className="w-6 h-6 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
          </div>
        ) : error ? (
          <div className="text-red-400 text-sm">{error}</div>
        ) : bids.length === 0 ? (
          <p className="text-zinc-500 text-sm">You have no bids.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-zinc-500 border-b border-white/10">
                  <th className="text-left py-2 font-medium">Bid ID</th>
                  <th className="text-left py-2 font-medium">Segment</th>
                  <th className="text-left py-2 font-medium">Credits</th>
                  <th className="text-left py-2 font-medium">Status</th>
                  <th className="text-left py-2 font-medium">Expires</th>
                  <th className="text-left py-2 font-medium">Created</th>
                </tr>
              </thead>
              <tbody>
                {bids.map((bid) => (
                  <tr
                    key={bid.id}
                    className="border-b border-white/5 last:border-0"
                  >
                    <td className="py-3 font-mono text-xs text-zinc-400">
                      {bid.id.slice(0, 8)}...
                    </td>
                    <td className="py-3 font-mono text-xs text-zinc-400">
                      {bid.segment_id.slice(0, 8)}...
                    </td>
                    <td className="py-3 font-medium">
                      {bid.bid_credits.toLocaleString()}
                    </td>
                    <td className="py-3">
                      <BidStatusBadge status={bid.status} />
                    </td>
                    <td className="py-3 text-zinc-500">
                      {new Date(bid.expires_at).toLocaleString()}
                    </td>
                    <td className="py-3 text-zinc-500">
                      {new Date(bid.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

function BidStatusBadge({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    active: "bg-blue-500/20 text-blue-400",
    accepted: "bg-emerald-500/20 text-emerald-400",
    rejected: "bg-red-500/20 text-red-400",
    expired: "bg-zinc-500/20 text-zinc-400",
    cancelled: "bg-amber-500/20 text-amber-400",
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

export default function PrivacyPage() {
  const { user } = useAuth();
  if (!user) return null;
  return user.role === "seller" ? <SellerPrivacyPage /> : <BuyerBidsPage />;
}
