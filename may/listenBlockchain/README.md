# listenBlockchain

Small Node.js example that subscribes to chain activity via [ethers v6](https://docs.ethers.org/v6/):

- **New blocks** — always on (HTTP polling via `JsonRpcProvider`).
- **ERC-20 `Transfer` events** — optional if you set `CONTRACT_ADDRESS` in `.env`.

## Setup

```bash
cd may/listenBlockchain
npm install
cp .env.example .env
```

Edit `.env`:

- `RPC_URL` — Sepolia (or any EVM) HTTPS RPC from [Alchemy](https://www.alchemy.com/), [Infura](https://www.infura.io/), etc.
- `CONTRACT_ADDRESS` — optional; any ERC-20 token on that network.

## Run

```bash
npm start
```

Example output:

```
Listening on RPC… Press Ctrl+C to stop.
2026-05-20T12:00:00.000Z new block 12345678
```

## Notes

- Do not commit `.env` (gitignored).
- For lower latency at scale, use a WebSocket RPC (`wss://`) with `ethers.WebSocketProvider`.
