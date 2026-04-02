"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/contexts/auth-context";
import {
  getDataset,
  purchaseDataset,
  type Dataset,
  type DatasetSample,
} from "@/lib/api";

export default function DatasetDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { user } = useAuth();

  const datasetId = params.id as string;

  const [dataset, setDataset] = useState<Dataset | null>(null);
  const [samples, setSamples] = useState<DatasetSample[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [purchasing, setPurchasing] = useState(false);
  const [purchaseError, setPurchaseError] = useState<string | null>(null);
  const [purchaseSuccess, setPurchaseSuccess] = useState(false);

  useEffect(() => {
    if (!datasetId) return;
    getDataset(datasetId)
      .then(({ dataset: d, samples: s }) => {
        setDataset(d);
        setSamples(s || []);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [datasetId]);

  async function handlePurchase() {
    if (!dataset) return;
    setPurchaseError(null);
    setPurchasing(true);

    try {
      await purchaseDataset(dataset.id);
      setPurchaseSuccess(true);
    } catch (err) {
      setPurchaseError(
        err instanceof Error ? err.message : "Purchase failed"
      );
    } finally {
      setPurchasing(false);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="w-8 h-8 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (error || !dataset) {
    return (
      <div className="space-y-4">
        <Link
          href="/marketplace"
          className="text-sm text-zinc-400 hover:text-white transition-colors"
        >
          &larr; Back to Marketplace
        </Link>
        <div className="bg-red-500/10 border border-red-500/20 text-red-400 rounded-xl p-6 text-center">
          <p className="font-medium mb-1">Dataset not found</p>
          <p className="text-sm">{error || "The requested dataset could not be loaded."}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Breadcrumb */}
      <Link
        href="/marketplace"
        className="inline-flex items-center text-sm text-zinc-400 hover:text-white transition-colors"
      >
        <svg
          className="w-4 h-4 mr-1"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          strokeWidth={1.5}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M15.75 19.5L8.25 12l7.5-7.5"
          />
        </svg>
        Back to Marketplace
      </Link>

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold mb-2">{dataset.title}</h1>
          {dataset.description && (
            <p className="text-zinc-400">{dataset.description}</p>
          )}
        </div>
        <div className="text-right shrink-0">
          <p className="text-3xl font-bold text-emerald-400">
            {dataset.current_price_credits}
          </p>
          <p className="text-xs text-zinc-500">credits</p>
          {dataset.base_price_credits !== dataset.current_price_credits && (
            <p className="text-xs text-zinc-600 line-through">
              {dataset.base_price_credits} base price
            </p>
          )}
        </div>
      </div>

      {/* Purchase actions */}
      {user?.role === "buyer" && (
        <div className="bg-white/5 border border-white/10 rounded-xl p-5">
          {purchaseSuccess ? (
            <div className="bg-emerald-500/10 border border-emerald-500/20 rounded-lg p-4 text-center">
              <p className="text-emerald-400 font-medium mb-1">
                Purchase successful!
              </p>
              <p className="text-sm text-zinc-400">
                The dataset has been added to your purchases.
              </p>
              <button
                onClick={() => router.push("/my-data")}
                className="mt-3 text-sm text-emerald-400 hover:underline"
              >
                View My Purchases
              </button>
            </div>
          ) : (
            <>
              {purchaseError && (
                <div className="bg-red-500/10 border border-red-500/20 text-red-400 text-sm rounded-lg p-3 mb-4">
                  {purchaseError}
                </div>
              )}
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-zinc-300">
                    Purchase this dataset for{" "}
                    <span className="font-bold text-emerald-400">
                      {dataset.current_price_credits} credits
                    </span>
                  </p>
                  <p className="text-xs text-zinc-500 mt-1">
                    Your balance: {user.credit_balance.toLocaleString()} credits
                  </p>
                </div>
                <button
                  onClick={handlePurchase}
                  disabled={
                    purchasing ||
                    user.credit_balance < dataset.current_price_credits
                  }
                  className="bg-emerald-500 hover:bg-emerald-400 disabled:opacity-50 disabled:cursor-not-allowed text-black font-semibold px-6 py-2.5 rounded-lg transition-colors text-sm"
                >
                  {purchasing
                    ? "Processing..."
                    : user.credit_balance < dataset.current_price_credits
                    ? "Insufficient credits"
                    : "Purchase Dataset"}
                </button>
              </div>
            </>
          )}
        </div>
      )}

      {/* Metadata grid */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
        <MetadataCard label="Contributors" value={String(dataset.contributor_count)} />
        <MetadataCard
          label="Records"
          value={dataset.record_count.toLocaleString()}
        />
        <MetadataCard label="K-Anonymity" value={`k = ${dataset.k_anonymity_k}`} />
        <MetadataCard
          label="Epsilon/Query"
          value={dataset.epsilon_per_query.toFixed(3)}
        />
        <MetadataCard label="Noise Mechanism" value={dataset.noise_mechanism} />
        <MetadataCard label="Status" value={dataset.status} />
        <MetadataCard
          label="Date Range"
          value={
            dataset.date_range_start
              ? `${new Date(dataset.date_range_start).toLocaleDateString()} - ${
                  dataset.date_range_end
                    ? new Date(dataset.date_range_end).toLocaleDateString()
                    : "..."
                }`
              : "N/A"
          }
        />
        <MetadataCard
          label="Created"
          value={new Date(dataset.created_at).toLocaleDateString()}
        />
      </div>

      {/* Categories */}
      {dataset.category_filter && dataset.category_filter.length > 0 && (
        <div className="bg-white/5 border border-white/10 rounded-xl p-5">
          <h2 className="text-sm font-semibold text-zinc-300 mb-3">
            Categories
          </h2>
          <div className="flex flex-wrap gap-2">
            {dataset.category_filter.map((cat) => (
              <span
                key={cat}
                className="text-xs bg-white/10 text-zinc-300 px-3 py-1 rounded-full"
              >
                {cat.replace(/_/g, " ")}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Demographics */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {dataset.age_ranges && dataset.age_ranges.length > 0 && (
          <div className="bg-white/5 border border-white/10 rounded-xl p-5">
            <h2 className="text-sm font-semibold text-zinc-300 mb-3">
              Age Ranges
            </h2>
            <div className="flex flex-wrap gap-2">
              {dataset.age_ranges.map((ar) => (
                <span
                  key={ar}
                  className="text-xs bg-blue-500/10 text-blue-400 px-3 py-1 rounded-full"
                >
                  {ar}
                </span>
              ))}
            </div>
          </div>
        )}
        {dataset.countries && dataset.countries.length > 0 && (
          <div className="bg-white/5 border border-white/10 rounded-xl p-5">
            <h2 className="text-sm font-semibold text-zinc-300 mb-3">
              Countries
            </h2>
            <div className="flex flex-wrap gap-2">
              {dataset.countries.map((c) => (
                <span
                  key={c}
                  className="text-xs bg-purple-500/10 text-purple-400 px-3 py-1 rounded-full"
                >
                  {c}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Sample data preview */}
      <div className="bg-white/5 border border-white/10 rounded-xl p-5">
        <h2 className="text-sm font-semibold text-zinc-300 mb-4">
          Sample Data Preview
        </h2>
        {samples.length === 0 ? (
          <p className="text-zinc-500 text-sm">
            No sample data available for preview.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-zinc-500 border-b border-white/10">
                  <th className="text-left py-2 font-medium">Category</th>
                  <th className="text-left py-2 font-medium">Duration Range</th>
                  <th className="text-left py-2 font-medium">Time of Day</th>
                  <th className="text-left py-2 font-medium">Device</th>
                  <th className="text-left py-2 font-medium">Age Range</th>
                  <th className="text-left py-2 font-medium">Country</th>
                </tr>
              </thead>
              <tbody>
                {samples.map((sample) => (
                  <tr
                    key={sample.id}
                    className="border-b border-white/5 last:border-0"
                  >
                    <td className="py-2.5">
                      <span className="text-xs bg-white/10 text-zinc-300 px-2 py-0.5 rounded-full">
                        {sample.app_category.replace(/_/g, " ")}
                      </span>
                    </td>
                    <td className="py-2.5 text-zinc-400">
                      {sample.duration_range}
                    </td>
                    <td className="py-2.5 text-zinc-400">
                      {sample.time_of_day}
                    </td>
                    <td className="py-2.5 text-zinc-500">
                      {sample.device_type || "--"}
                    </td>
                    <td className="py-2.5 text-zinc-500">
                      {sample.contributor_age_range || "--"}
                    </td>
                    <td className="py-2.5 text-zinc-500">
                      {sample.contributor_country || "--"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        <p className="text-xs text-zinc-600 mt-3">
          Showing anonymized sample records. Actual data contains generalized
          values to protect contributor privacy.
        </p>
      </div>
    </div>
  );
}

function MetadataCard({
  label,
  value,
}: {
  label: string;
  value: string;
}) {
  return (
    <div className="bg-white/5 border border-white/10 rounded-lg p-4">
      <p className="text-xs text-zinc-500 uppercase tracking-wider mb-1">
        {label}
      </p>
      <p className="text-sm font-medium text-zinc-200">{value}</p>
    </div>
  );
}
