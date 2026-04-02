"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { AuthProvider, useAuth } from "@/contexts/auth-context";
import type { UserRole } from "@/lib/api";

const AGE_RANGES = ["13-17", "18-24", "25-34", "35-44", "45-54", "55-64", "65+"];

const COUNTRIES = [
  "US",
  "GB",
  "CA",
  "AU",
  "DE",
  "FR",
  "JP",
  "IN",
  "BR",
  "Other",
];

function RegisterForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { register } = useAuth();

  const defaultRole = (searchParams.get("role") as UserRole) || "seller";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [role, setRole] = useState<UserRole>(defaultRole);
  const [ageRange, setAgeRange] = useState("");
  const [country, setCountry] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      await register({
        email,
        password,
        display_name: displayName,
        role,
        age_range: ageRange || undefined,
        country: country || undefined,
      });
      router.push("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <Link href="/" className="text-2xl font-bold tracking-tight">
            Screener
          </Link>
          <p className="text-zinc-500 mt-2">Create your account</p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="bg-white/5 border border-white/10 rounded-2xl p-8 space-y-5"
        >
          {error && (
            <div className="bg-red-500/10 border border-red-500/20 text-red-400 text-sm rounded-lg p-3">
              {error}
            </div>
          )}

          <div>
            <label
              htmlFor="displayName"
              className="block text-sm font-medium text-zinc-300 mb-2"
            >
              Display Name
            </label>
            <input
              id="displayName"
              type="text"
              required
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500/50"
              placeholder="Your name"
            />
          </div>

          <div>
            <label
              htmlFor="email"
              className="block text-sm font-medium text-zinc-300 mb-2"
            >
              Email
            </label>
            <input
              id="email"
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500/50"
              placeholder="you@example.com"
            />
          </div>

          <div>
            <label
              htmlFor="password"
              className="block text-sm font-medium text-zinc-300 mb-2"
            >
              Password
            </label>
            <input
              id="password"
              type="password"
              required
              minLength={8}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500/50"
              placeholder="Min. 8 characters"
            />
          </div>

          {/* Role selector */}
          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-2">
              I want to...
            </label>
            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={() => setRole("seller")}
                className={`py-3 rounded-lg border text-sm font-medium transition-colors ${
                  role === "seller"
                    ? "bg-emerald-500/20 border-emerald-500/50 text-emerald-400"
                    : "bg-white/5 border-white/10 text-zinc-400 hover:border-white/20"
                }`}
              >
                Sell my data
              </button>
              <button
                type="button"
                onClick={() => setRole("buyer")}
                className={`py-3 rounded-lg border text-sm font-medium transition-colors ${
                  role === "buyer"
                    ? "bg-blue-500/20 border-blue-500/50 text-blue-400"
                    : "bg-white/5 border-white/10 text-zinc-400 hover:border-white/20"
                }`}
              >
                Buy datasets
              </button>
            </div>
          </div>

          {/* Optional fields */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="ageRange"
                className="block text-sm font-medium text-zinc-300 mb-2"
              >
                Age Range{" "}
                <span className="text-zinc-600">(optional)</span>
              </label>
              <select
                id="ageRange"
                value={ageRange}
                onChange={(e) => setAgeRange(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500/50"
              >
                <option value="" className="bg-zinc-900">
                  Select
                </option>
                {AGE_RANGES.map((a) => (
                  <option key={a} value={a} className="bg-zinc-900">
                    {a}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label
                htmlFor="country"
                className="block text-sm font-medium text-zinc-300 mb-2"
              >
                Country{" "}
                <span className="text-zinc-600">(optional)</span>
              </label>
              <select
                id="country"
                value={country}
                onChange={(e) => setCountry(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500/50"
              >
                <option value="" className="bg-zinc-900">
                  Select
                </option>
                {COUNTRIES.map((c) => (
                  <option key={c} value={c} className="bg-zinc-900">
                    {c}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-emerald-500 hover:bg-emerald-400 disabled:opacity-50 disabled:cursor-not-allowed text-black font-semibold py-3 rounded-lg transition-colors"
          >
            {loading ? "Creating account..." : "Create account"}
          </button>
        </form>

        <p className="text-center text-sm text-zinc-500 mt-6">
          Already have an account?{" "}
          <Link href="/login" className="text-emerald-400 hover:underline">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}

export default function RegisterPage() {
  return (
    <AuthProvider>
      <RegisterForm />
    </AuthProvider>
  );
}
