"use client";

import { useEffect, useState } from "react";
import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { LAMPORTS_PER_SOL } from "@solana/web3.js";
import { useAuth } from "@/contexts/auth-context";
import { getSolTransactions, getSolanaInfo, type SolTransaction, type SolServerInfo } from "@/lib/api";
import { SolTopup } from "@/components/solana/SolTopup";

export default function SolanaPage() {
  const { user } = useAuth();
  const { publicKey, connected } = useWallet();
  const { connection } = useConnection();

  const [solBalance, setSolBalance] = useState<number | null>(null);
  const [serverInfo, setServerInfo] = useState<SolServerInfo | null>(null);
  const [transactions, setTransactions] = useState<SolTransaction[]>([]);
  const [totalTx, setTotalTx] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getSolanaInfo()
      .then(setServerInfo)
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!connected || !publicKey) {
      setSolBalance(null);
      return;
    }
    connection
      .getBalance(publicKey)
      .then(setSolBalance)
      .catch(() => setSolBalance(null));
  }, [connected, publicKey, connection]);

  useEffect(() => {
    setLoading(true);
    getSolTransactions(50, 0)
      .then(({ transactions: txs, total }) => {
        setTransactions(txs);
        setTotalTx(total);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const txTypeLabel: Record<string, string> = {
    topup: "Credit Top-Up",
    purchase: "Dataset Purchase",
    seller_payout: "Seller Payout",
    escrow_deposit: "Escrow Deposit",
    escrow_release: "Escrow Release",
    escrow_refund: "Escrow Refund",
  };

  const statusColor: Record<string, string> = {
    confirmed: "text-emerald-400",
    pending: "text-yellow-400",
    failed: "text-red-400",
  };

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold mb-1">Solana</h1>
        <p className="text-zinc-400 text-sm">
          Manage your Solana wallet, top up credits with SOL, and view transaction history.
        </p>
      </div>

      {/* Wallet Status */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-white/5 border border-white/10 rounded-xl p-5">
          <p className="text-xs text-zinc-500 uppercase tracking-wider mb-1">
            Wallet Status
          </p>
          {connected && publicKey ? (
            <div>
              <p className="text-sm font-medium text-emerald-400">Connected</p>
              <p className="text-xs text-zinc-500 mt-1 font-mono">
                {publicKey.toBase58().slice(0, 8)}...{publicKey.toBase58().slice(-8)}
              </p>
            </div>
          ) : (
            <p className="text-sm text-zinc-500">Not connected</p>
          )}
        </div>

        <div className="bg-white/5 border border-white/10 rounded-xl p-5">
          <p className="text-xs text-zinc-500 uppercase tracking-wider mb-1">
            SOL Balance
          </p>
          {solBalance !== null ? (
            <p className="text-lg font-bold text-purple-400">
              {(solBalance / LAMPORTS_PER_SOL).toFixed(4)} SOL
            </p>
          ) : (
            <p className="text-sm text-zinc-500">--</p>
          )}
        </div>

        <div className="bg-white/5 border border-white/10 rounded-xl p-5">
          <p className="text-xs text-zinc-500 uppercase tracking-wider mb-1">
            Linked Wallet
          </p>
          {user?.solana_wallet ? (
            <div>
              <p className="text-sm font-medium text-emerald-400">Linked</p>
              <p className="text-xs text-zinc-500 mt-1 font-mono">
                {user.solana_wallet.slice(0, 8)}...{user.solana_wallet.slice(-8)}
              </p>
            </div>
          ) : (
            <p className="text-sm text-zinc-500">Not linked</p>
          )}
        </div>
      </div>

      {/* Server info */}
      {serverInfo && (
        <div className="bg-white/5 border border-white/10 rounded-xl p-5">
          <h2 className="text-sm font-semibold text-zinc-300 mb-3">
            Network Info
          </h2>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
            <div>
              <p className="text-xs text-zinc-500">Network</p>
              <p className="text-zinc-300">Devnet</p>
            </div>
            <div>
              <p className="text-xs text-zinc-500">Exchange Rate</p>
              <p className="text-zinc-300">
                {serverInfo.lamports_per_credit.toLocaleString()} lamports/credit
              </p>
            </div>
            <div>
              <p className="text-xs text-zinc-500">Program ID</p>
              <p className="text-zinc-300 font-mono text-xs truncate">
                {serverInfo.program_id.slice(0, 12)}...
              </p>
            </div>
            <div>
              <p className="text-xs text-zinc-500">Server Balance</p>
              <p className="text-zinc-300">
                {(serverInfo.sol_balance_lamports / LAMPORTS_PER_SOL).toFixed(2)} SOL
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Top-up */}
      <SolTopup />

      {/* Transaction History */}
      <div className="bg-white/5 border border-white/10 rounded-xl p-5">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-sm font-semibold text-zinc-300">
            Transaction History
          </h2>
          <span className="text-xs text-zinc-500">{totalTx} total</span>
        </div>

        {loading ? (
          <div className="flex justify-center py-8">
            <div className="w-6 h-6 border-2 border-purple-500 border-t-transparent rounded-full animate-spin" />
          </div>
        ) : transactions.length === 0 ? (
          <p className="text-zinc-500 text-sm text-center py-8">
            No Solana transactions yet.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-zinc-500 border-b border-white/10">
                  <th className="text-left py-2 font-medium">Type</th>
                  <th className="text-left py-2 font-medium">Amount</th>
                  <th className="text-left py-2 font-medium">Status</th>
                  <th className="text-left py-2 font-medium">Date</th>
                  <th className="text-left py-2 font-medium">Tx</th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((tx) => (
                  <tr
                    key={tx.id}
                    className="border-b border-white/5 last:border-0"
                  >
                    <td className="py-2.5 text-zinc-300">
                      {txTypeLabel[tx.tx_type] || tx.tx_type}
                    </td>
                    <td className="py-2.5 text-zinc-300 font-mono">
                      {(tx.amount_lamports / LAMPORTS_PER_SOL).toFixed(4)} SOL
                    </td>
                    <td className={`py-2.5 ${statusColor[tx.status] || "text-zinc-400"}`}>
                      {tx.status}
                    </td>
                    <td className="py-2.5 text-zinc-500">
                      {new Date(tx.created_at).toLocaleDateString()}
                    </td>
                    <td className="py-2.5">
                      <a
                        href={`https://explorer.solana.com/tx/${tx.tx_signature}?cluster=devnet`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-purple-400 hover:text-purple-300 text-xs font-mono"
                      >
                        {tx.tx_signature.slice(0, 8)}...
                      </a>
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
