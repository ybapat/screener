"use client";

import { useEffect, useState, useCallback } from "react";
import { useAuth } from "@/contexts/auth-context";
import {
  uploadData,
  getBatches,
  getPurchases,
  type DataBatch,
  type Purchase,
  type ScreenTimeRecordInput,
} from "@/lib/api";

// ---------------------------------------------------------------------------
// Seller: upload data and view batches
// ---------------------------------------------------------------------------

function SellerDataPage() {
  const [batches, setBatches] = useState<DataBatch[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Upload state
  const [jsonInput, setJsonInput] = useState("");
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [uploadSuccess, setUploadSuccess] = useState<string | null>(null);

  const loadBatches = useCallback(async () => {
    try {
      const result = await getBatches();
      setBatches(result.batches);
      setTotal(result.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load batches");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadBatches();
  }, [loadBatches]);

  async function handleUpload(e: React.FormEvent) {
    e.preventDefault();
    setUploadError(null);
    setUploadSuccess(null);
    setUploading(true);

    try {
      const records: ScreenTimeRecordInput[] = JSON.parse(jsonInput);

      if (!Array.isArray(records) || records.length === 0) {
        throw new Error("Input must be a non-empty JSON array of records");
      }

      const batch = await uploadData(records);
      setUploadSuccess(
        `Uploaded ${batch.record_count} records (Batch ${batch.id.slice(0, 8)})`
      );
      setJsonInput("");
      loadBatches();
    } catch (err) {
      setUploadError(
        err instanceof Error ? err.message : "Upload failed"
      );
    } finally {
      setUploading(false);
    }
  }

  function loadSampleData() {
    const sample: ScreenTimeRecordInput[] = [
      {
        app_name: "Instagram",
        app_category: "social_media",
        duration_secs: 1800,
        started_at: new Date(Date.now() - 7200000).toISOString(),
        ended_at: new Date(Date.now() - 5400000).toISOString(),
        device_type: "phone",
        os: "ios",
      },
      {
        app_name: "YouTube",
        app_category: "entertainment",
        duration_secs: 3600,
        started_at: new Date(Date.now() - 14400000).toISOString(),
        ended_at: new Date(Date.now() - 10800000).toISOString(),
        device_type: "tablet",
        os: "android",
      },
      {
        app_name: "VS Code",
        app_category: "productivity",
        duration_secs: 5400,
        started_at: new Date(Date.now() - 21600000).toISOString(),
        ended_at: new Date(Date.now() - 16200000).toISOString(),
        device_type: "desktop",
        os: "macos",
      },
    ];
    setJsonInput(JSON.stringify(sample, null, 2));
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold mb-1">My Data</h1>
        <p className="text-zinc-500">Upload screen time data and track your batches</p>
      </div>

      {/* Upload form */}
      <div className="bg-white/5 border border-white/10 rounded-xl p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-sm font-semibold text-zinc-300">Upload Data</h2>
          <button
            type="button"
            onClick={loadSampleData}
            className="text-xs text-emerald-400 hover:text-emerald-300 transition-colors"
          >
            Load sample data
          </button>
        </div>

        <form onSubmit={handleUpload} className="space-y-4">
          {uploadError && (
            <div className="bg-red-500/10 border border-red-500/20 text-red-400 text-sm rounded-lg p-3">
              {uploadError}
            </div>
          )}
          {uploadSuccess && (
            <div className="bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-sm rounded-lg p-3">
              {uploadSuccess}
            </div>
          )}

          <textarea
            value={jsonInput}
            onChange={(e) => setJsonInput(e.target.value)}
            rows={12}
            className="w-full bg-black/30 border border-white/10 rounded-lg px-4 py-3 text-sm font-mono text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 resize-y"
            placeholder='[{"app_name": "Instagram", "app_category": "social_media", "duration_secs": 1800, "started_at": "2025-01-01T10:00:00Z", "ended_at": "2025-01-01T10:30:00Z"}]'
          />

          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={uploading || !jsonInput.trim()}
              className="bg-emerald-500 hover:bg-emerald-400 disabled:opacity-50 disabled:cursor-not-allowed text-black font-semibold px-6 py-2.5 rounded-lg transition-colors text-sm"
            >
              {uploading ? "Uploading..." : "Upload Records"}
            </button>
            <p className="text-xs text-zinc-500">
              Paste a JSON array of screen time records (max 1,000 per batch)
            </p>
          </div>
        </form>
      </div>

      {/* Batches list */}
      <div className="bg-white/5 border border-white/10 rounded-xl p-6">
        <h2 className="text-sm font-semibold text-zinc-300 mb-4">
          Uploaded Batches ({total})
        </h2>

        {loading ? (
          <div className="flex items-center justify-center py-10">
            <div className="w-6 h-6 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
          </div>
        ) : error ? (
          <div className="text-red-400 text-sm">{error}</div>
        ) : batches.length === 0 ? (
          <p className="text-zinc-500 text-sm">
            No batches uploaded yet. Upload your first batch above.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-zinc-500 border-b border-white/10">
                  <th className="text-left py-2 font-medium">Batch ID</th>
                  <th className="text-left py-2 font-medium">Records</th>
                  <th className="text-left py-2 font-medium">Date Range</th>
                  <th className="text-left py-2 font-medium">Status</th>
                  <th className="text-left py-2 font-medium">Created</th>
                </tr>
              </thead>
              <tbody>
                {batches.map((batch) => (
                  <tr
                    key={batch.id}
                    className="border-b border-white/5 last:border-0"
                  >
                    <td className="py-3 font-mono text-xs text-zinc-400">
                      {batch.id.slice(0, 8)}...
                    </td>
                    <td className="py-3">{batch.record_count}</td>
                    <td className="py-3 text-zinc-400 text-xs">
                      {batch.date_range_start
                        ? `${new Date(batch.date_range_start).toLocaleDateString()} - ${
                            batch.date_range_end
                              ? new Date(batch.date_range_end).toLocaleDateString()
                              : "..."
                          }`
                        : "--"}
                    </td>
                    <td className="py-3">
                      <StatusBadge status={batch.status} />
                    </td>
                    <td className="py-3 text-zinc-500">
                      {new Date(batch.created_at).toLocaleDateString()}
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

// ---------------------------------------------------------------------------
// Buyer: view purchases
// ---------------------------------------------------------------------------

function BuyerPurchasesPage() {
  const [purchases, setPurchases] = useState<Purchase[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getPurchases()
      .then(({ purchases: p, total: t }) => {
        setPurchases(p);
        setTotal(t);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold mb-1">My Purchases</h1>
        <p className="text-zinc-500">Datasets you have purchased</p>
      </div>

      <div className="bg-white/5 border border-white/10 rounded-xl p-6">
        <h2 className="text-sm font-semibold text-zinc-300 mb-4">
          All Purchases ({total})
        </h2>

        {loading ? (
          <div className="flex items-center justify-center py-10">
            <div className="w-6 h-6 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
          </div>
        ) : error ? (
          <div className="text-red-400 text-sm">{error}</div>
        ) : purchases.length === 0 ? (
          <p className="text-zinc-500 text-sm">
            You have not purchased any datasets yet.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-zinc-500 border-b border-white/10">
                  <th className="text-left py-2 font-medium">Purchase ID</th>
                  <th className="text-left py-2 font-medium">Dataset</th>
                  <th className="text-left py-2 font-medium">Price</th>
                  <th className="text-left py-2 font-medium">Status</th>
                  <th className="text-left py-2 font-medium">Downloads</th>
                  <th className="text-left py-2 font-medium">Date</th>
                </tr>
              </thead>
              <tbody>
                {purchases.map((p) => (
                  <tr
                    key={p.id}
                    className="border-b border-white/5 last:border-0"
                  >
                    <td className="py-3 font-mono text-xs text-zinc-400">
                      {p.id.slice(0, 8)}...
                    </td>
                    <td className="py-3 font-mono text-xs text-zinc-400">
                      {p.dataset_id.slice(0, 8)}...
                    </td>
                    <td className="py-3">{p.price_credits} credits</td>
                    <td className="py-3">
                      <StatusBadge status={p.status} />
                    </td>
                    <td className="py-3">{p.download_count}</td>
                    <td className="py-3 text-zinc-500">
                      {new Date(p.purchased_at).toLocaleDateString()}
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

export default function MyDataPage() {
  const { user } = useAuth();
  if (!user) return null;
  return user.role === "seller" ? <SellerDataPage /> : <BuyerPurchasesPage />;
}
