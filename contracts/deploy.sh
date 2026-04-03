#!/usr/bin/env bash
set -euo pipefail

# Deploy the Screener escrow program to Solana devnet.
#
# Prerequisites:
#   - solana CLI installed (or use Dockerfile.anchor to build)
#   - anchor CLI installed
#   - A funded devnet wallet at ~/.config/solana/id.json
#
# Usage:
#   cd contracts
#   ./deploy.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/escrow"

echo "==> Configuring Solana CLI for devnet..."
solana config set --url https://api.devnet.solana.com

echo "==> Requesting airdrop for deployment fees..."
solana airdrop 2 || echo "Airdrop may have failed (rate limited). Ensure wallet is funded."

echo "==> Building program..."
anchor build

echo "==> Deploying to devnet..."
anchor deploy --provider.cluster devnet

PROGRAM_ID=$(solana-keygen pubkey target/deploy/screener_escrow-keypair.json)
echo ""
echo "==> Deployed! Program ID: $PROGRAM_ID"
echo ""
echo "Update the following with this program ID:"
echo "  - contracts/escrow/Anchor.toml"
echo "  - contracts/escrow/programs/escrow/src/lib.rs (declare_id!)"
echo "  - backend env: SOLANA_PROGRAM_ID=$PROGRAM_ID"
