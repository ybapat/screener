"use client";

import { useState } from "react";
import { useWallet } from "@solana/wallet-adapter-react";
import { WalletMultiButton } from "@solana/wallet-adapter-react-ui";
import { useAuth } from "@/contexts/auth-context";
import { linkSolanaWallet } from "@/lib/api";

export function WalletButton() {
  const { publicKey, signMessage, connected } = useWallet();
  const { user, refreshUser } = useAuth();
  const [linking, setLinking] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleLinkWallet() {
    if (!publicKey || !signMessage) return;
    setLinking(true);
    setError(null);

    try {
      const message = `Link wallet ${publicKey.toBase58()} to Screener account`;
      const encoded = new TextEncoder().encode(message);
      const signature = await signMessage(encoded);
      const sigB64 = Buffer.from(signature).toString("base64");

      await linkSolanaWallet(publicKey.toBase58(), sigB64, message);
      await refreshUser();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to link wallet");
    } finally {
      setLinking(false);
    }
  }

  return (
    <div className="flex flex-col gap-2">
      <WalletMultiButton className="!bg-purple-600 hover:!bg-purple-700 !rounded-lg !text-sm !h-9" />

      {connected && publicKey && !user?.solana_wallet && (
        <button
          onClick={handleLinkWallet}
          disabled={linking}
          className="text-xs bg-purple-600/20 text-purple-400 hover:bg-purple-600/30 px-3 py-1.5 rounded-lg transition-colors disabled:opacity-50"
        >
          {linking ? "Linking..." : "Link Wallet to Account"}
        </button>
      )}

      {user?.solana_wallet && (
        <p className="text-xs text-zinc-500">
          Linked: {user.solana_wallet.slice(0, 4)}...
          {user.solana_wallet.slice(-4)}
        </p>
      )}

      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  );
}
