# Contributing to KYRELIS

Thank you for contributing to **KYRELIS**, a student-built Cosmos SDK blockchain and its developer tools. Issues, documentation corrections, interoperability testing, and focused pull requests are welcome.

## Before you start

- Read [README.md](README.md), [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md), and [SECURITY.md](SECURITY.md).
- Search existing issues and pull requests to avoid duplicates.
- For small bugs, open a bug report; for new capabilities, use the feature request form.
- For significant protocol, tokenomics, genesis, staking, fee, wallet-signing, or security changes, discuss the design in an issue **before** coding.
- Never disclose keys, seeds, environment files, local node state, real account data, or sensitive endpoints.

## Repository layout

| Path | Purpose |
| --- | --- |
| `chain/` | Cosmos SDK / CometBFT chain, application logic, CLI and tests |
| `hub/` | Django service and Vite/Keplr web wallet |
| `operator/` | Go-based operator interface and RPC integration |
| `scripts/` | Local network bootstrap and Hub launch tools |
| `START*.cmd`, `start*.sh` | Windows and Unix launchers |
| `media/` | Project imagery |

## Development environment

The README is the source of truth for prerequisites and startup. The chain currently specifies Go **1.26.5**, the operator specifies Go **1.23**, and the web Hub requires Node.js **22.12+**; the local runner also requires Python **3.10+**. Different modules intentionally have separate requirements.

### Fast checks

```sh
(cd operator && go test ./...)
(cd hub/web && npm ci && npm run build)
python3 -m py_compile scripts/devnet.py scripts/hub.py hub/server.py
```

For chain changes, after installing the required Go toolchain and downloading dependencies:

```sh
(cd chain && go test ./app ./cmd/kyrelisd/cmd)
```

For full chain testing, follow the current module's test requirements; Cosmos simulation tests may be resource-intensive. Avoid using production credentials or mainnet funds. For platform-specific scripts, also test the relevant `.cmd` on Windows or `.sh` on Linux.

### Keplr and transaction changes

Do not treat a displayed balance as proof of compatibility. Test the intended chain ID and Bech32 prefix (`kyr`), fee denom (`ukyrl`), chain/account sequence, signature mode, gas estimation, broadcast response, transaction inclusion, and final balance. Record reproducible **redacted** failures.

## Pull request workflow

1. Fork the repository and create a narrowly scoped branch.
2. Add or adjust tests and documentation for behavior changes.
3. Run applicable checks; explain any checks you could not run.
4. Open a pull request using the template and link an issue.
5. Respond constructively to feedback; maintainers decide whether and when changes are merged.

Favor small, reviewable commits and clear commit messages (`fix:`, `feat:`, `docs:`, `test:`, `chore:`). Format code according to its language conventions. Preserve existing third-party notices in `chain/LICENSE-Ignite` and `operator/LICENSE`.

## High-risk changes

Consensus, genesis supply, transaction validation, key custody, address formats, fees, and upgrades require explicit review, a reproducible test network, compatibility notes, and a rollback/migration plan. A passing CI job is not a security audit or permission to operate with real funds.

See [COLLABORATION.md](COLLABORATION.md) for roles, proposals and review expectations.
