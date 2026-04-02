"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { getDatasets, type Dataset } from "@/lib/api";

const CATEGORIES = [
  "social_media",
  "entertainment",
  "productivity",
  "gaming",
  "education",
  "health",
  "finance",
  "communication",
  "news",
  "shopping",
];

export default function MarketplacePage() {
  const [datasets, setDatasets] = useState<Dataset[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);

  const loadDatasets = useCallback(async () => {
    setLoading(true);
    try {
      const result = await getDatasets(
        selectedCategories.length > 0 ? selectedCategories : undefined,
        20,
        0
      );
      setDatasets(result.datasets);
      setTotal(result.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load datasets");
    } finally {
      setLoading(false);
    }
  }, [selectedCategories]);

  useEffect(() => {
    loadDatasets();
  }, [loadDatasets]);

  function toggleCategory(cat: string) {
    setSelectedCategories((prev) =>
      prev.includes(cat) ? prev.filter((c) => c !== cat) : [...prev, cat]
    );
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold mb-1">Marketplace</h1>
        <p className="text-zinc-500">
          Browse anonymized screen time datasets ({total} available)
        </p>
      </div>

      {/* Category filters */}
      <div className="flex flex-wrap gap-2">
        {CATEGORIES.map((cat) => {
          const isActive = selectedCategories.includes(cat);
          return (
            <button
              key={cat}
              onClick={() => toggleCategory(cat)}
              className={`text-xs px-3 py-1.5 rounded-full border transition-colors ${
                isActive
                  ? "bg-emerald-500/20 border-emerald-500/50 text-emerald-400"
                  : "bg-white/5 border-white/10 text-zinc-400 hover:border-white/20 hover:text-zinc-300"
              }`}
            >
              {cat.replace(/_/g, " ")}
            </button>
          );
        })}
        {selectedCategories.length > 0 && (
          <button
            onClick={() => setSelectedCategories([])}
            className="text-xs px-3 py-1.5 rounded-full border border-red-500/30 text-red-400 hover:bg-red-500/10 transition-colors"
          >
            Clear filters
          </button>
        )}
      </div>

      {/* Dataset grid */}
      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="w-8 h-8 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
        </div>
      ) : error ? (
        <div className="bg-red-500/10 border border-red-500/20 text-red-400 rounded-xl p-6 text-center">
          <p className="font-medium mb-1">Failed to load datasets</p>
          <p className="text-sm">{error}</p>
        </div>
      ) : datasets.length === 0 ? (
        <div className="text-center py-20 text-zinc-500">
          <p className="text-lg mb-2">No datasets found</p>
          <p className="text-sm">
            {selectedCategories.length > 0
              ? "Try clearing your filters"
              : "Check back later for new datasets"}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {datasets.map((dataset) => (
            <DatasetCard key={dataset.id} dataset={dataset} />
          ))}
        </div>
      )}
    </div>
  );
}

function DatasetCard({ dataset }: { dataset: Dataset }) {
  return (
    <Link
      href={`/marketplace/${dataset.id}`}
      className="block bg-white/5 border border-white/10 rounded-xl p-5 hover:border-white/20 hover:bg-white/[0.07] transition-all group"
    >
      <div className="flex items-start justify-between mb-3">
        <h3 className="font-semibold text-sm group-hover:text-emerald-400 transition-colors line-clamp-1">
          {dataset.title}
        </h3>
        <span className="text-emerald-400 font-bold text-sm shrink-0 ml-2">
          {dataset.current_price_credits} cr
        </span>
      </div>

      {dataset.description && (
        <p className="text-xs text-zinc-500 mb-3 line-clamp-2">
          {dataset.description}
        </p>
      )}

      {/* Category tags */}
      <div className="flex flex-wrap gap-1 mb-3">
        {(dataset.category_filter || []).slice(0, 3).map((cat) => (
          <span
            key={cat}
            className="text-[10px] bg-white/10 text-zinc-400 px-2 py-0.5 rounded-full"
          >
            {cat.replace(/_/g, " ")}
          </span>
        ))}
        {(dataset.category_filter || []).length > 3 && (
          <span className="text-[10px] text-zinc-600">
            +{dataset.category_filter.length - 3}
          </span>
        )}
      </div>

      {/* Metadata row */}
      <div className="flex items-center gap-4 text-xs text-zinc-500">
        <span className="flex items-center gap-1">
          <svg
            className="w-3.5 h-3.5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            strokeWidth={1.5}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-1.053M18 10.5a3 3 0 11-6 0 3 3 0 016 0z"
            />
          </svg>
          {dataset.contributor_count} contributors
        </span>
        <span>{dataset.record_count.toLocaleString()} records</span>
        <span className="flex items-center gap-1">
          <svg
            className="w-3.5 h-3.5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            strokeWidth={1.5}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"
            />
          </svg>
          k={dataset.k_anonymity_k}
        </span>
      </div>

      {/* Status */}
      <div className="mt-3 pt-3 border-t border-white/5 flex items-center justify-between">
        <span
          className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${
            dataset.status === "active"
              ? "bg-emerald-500/20 text-emerald-400"
              : "bg-zinc-500/20 text-zinc-400"
          }`}
        >
          {dataset.status}
        </span>
        <span className="text-[10px] text-zinc-600">
          {new Date(dataset.created_at).toLocaleDateString()}
        </span>
      </div>
    </Link>
  );
}
