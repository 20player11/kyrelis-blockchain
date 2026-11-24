# Support and troubleshooting

KYRELIS is an educational project and support is community/maintainer best-effort, not a managed production service.

## Start here

1. Read [README.md](README.md) for installation, supported operating systems and launching local components.
2. Check existing issues for your exact error and component.
3. Ask a non-sensitive usage question in [Discussions](https://github.com/20player11/kyrelis-blockchain/discussions) if enabled, or submit a reproducible bug report if something is broken.
4. Follow [SECURITY.md](SECURITY.md) for vulnerabilities; never disclose secrets in issues.

## Useful diagnostics

- Check the local chain with `python3 scripts/devnet.py status` (Windows: `py -3 scripts/devnet.py status`).
- Check node RPC at `http://127.0.0.1:26657/status` when running a local node.
- Record the chain ID and revision, operating system, Go/Python/Node versions, and sanitized error output.
- If a Keplr balance appears but a transaction fails, check the chain ID, Bech32 `kyr` prefix, account existence/sequence, signer, fee denom `ukyrl`, RPC/REST availability and transaction broadcast/confirmation.
- If SSH/network access fails, distinguish local routing, VPN/firewall issues and service availability before changing the node.

## What not to share

Never share mnemonics, private keys, validator/node identity keys, authentication cookies, tokens, raw `.env` files, or unredacted node state. Screenshots and logs must be sanitized.

## Compatibility

Development chains may be recreated and existing state may not migrate automatically. The locally generated network is **not automatically** an independently hosted public KYRELIS mainnet or a wallet-registry listing.
