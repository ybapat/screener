"use client";

import { useState } from "react";
import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { PublicKey, SystemProgram, Transaction } from "@solana/web3.js";
import { initSolTopup, confirmSolTopup } from "@/lib/api";
import { useAuth } from "@/contexts/auth-context";

export function SolTopup() {
  const { connection } = useConnection();
  const { publicKey, sendTransaction, connected } = useWallet();
  const { user, refreshUser } = useAuth();
  const [amount, setAmount] = useState("");
  const [status, setStatus] = useState<
    "idle" | "signing" | "confirming" | "done" | "error"
  >("idle");
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<{
    creditsAdded: number;
    txSig: string;
  } | null>(null);

  async function handleTopup() {
    if (!publicKey || !sendTransaction || !amount) return;
    setStatus("signing");
    setError(null);
    setResult(null);

    try {
      const lamports = Math.floor(parseFloat(amount) * 1_000_000_000);
      if (lamports <= 0) throw new Error("Amount must be greater than 0");

      // 1. Get server wallet address
      const init = await initSolTopup(lamports);

      // 2. Build transfer transaction
      const tx = new Transaction().add(
        SystemProgram.transfer({
          fromPubkey: publicKey,
          toPubkey: new PublicKey(init.recipient_wallet),
          lamports: init.amount_lamports,
        })
      );

      // 3. Sign and send via wallet
      const signature = await sendTransaction(tx, connection);
      setStatus("confirming");

      // 4. Wait for on-chain confirmation
      await connection.confirmTransaction(signature, "confirmed");

      // 5. Tell backend to verify and credit
      const resp = await confirmSolTopup(signature);
      setStatus("done");
      setResult({
        creditsAdded: resp.credits_added,
        txSig: signature,
      });
      await refreshUser();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Transaction failed");
      setStatus("error");
    }
  }

  if (!connected || !user?.solana_wallet) {
    return null;
  }

  return (
    <div className="bg-white/5 border border-white/10 rounded-xl p-6">
      <h3 className="text-lg font-semibold text-white mb-4">
        Top Up with SOL
      </h3>

      <div className="flex gap-3 mb-4">
        <input
          type="number"
          step="0.001"
          min="0.001"
          placeholder="Amount in SOL"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          className="flex-1 bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white placeholder-zinc-500 focus:border-purple-500 focus:outline-none"
          disabled={status === "signing" || status === "confirming"}
        />
        <button
          onClick={handleTopup}
          disabled={
            !amount ||
            status === "signing" ||
            status === "confirming" ||
            parseFloat(amount) <= 0
          }
          className="bg-purple-600 hover:bg-purple-700 disabled:opacity-50 disabled:hover:bg-purple-600 text-white px-6 py-2 rounded-lg transition-colors font-medium"
        >
          {status === "signing"
            ? "Sign in Wallet..."
            : status === "confirming"
            ? "Confirming..."
            : "Top Up"}
        </button>
      </div>

      {amount && parseFloat(amount) > 0 && (
        <p className="text-sm text-zinc-400 mb-3">
          ≈ {Math.floor(parseFloat(amount) * 1_000_000_000 / 10000).toLocaleString()} credits
        </p>
      )}

      {status === "done" && result && (
        <div className="bg-green-500/10 border border-green-500/20 rounded-lg p-4">
          <p className="text-green-400 text-sm font-medium">
            +{result.creditsAdded.toLocaleString()} credits added
          </p>
          <a
            href={`https://explorer.solana.com/tx/${result.txSig}?cluster=devnet`}
            target="_blank"
            rel="noopener noreferrer"
            className="text-xs text-purple-400 hover:text-purple-300 mt-1 inline-block"
          >
            View on Solana Explorer →
          </a>
        </div>
      )}

      {status === "error" && error && (
        <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
          <p className="text-red-400 text-sm">{error}</p>
          <button
            onClick={() => {
              setStatus("idle");
              setError(null);
            }}
            className="text-xs text-zinc-400 hover:text-white mt-2"
          >
            Try again
          </button>
        </div>
      )}
    </div>
  );
}
