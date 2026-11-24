# KYRELIS security policy

## Project status and scope

KYRELIS is a **development / educational** Cosmos SDK blockchain with a local web Hub and operator interface. No component should be considered independently audited, production-hardened, or safe for real-value assets. Local test keyrings, generated genesis accounts and node-state directories are **not** secure custody solutions.

## Supported versions

| Version / branch | Security fixes |
| --- | --- |
| `main` (current development) | Best-effort review |
| Older commits, forks and custom networks | Not guaranteed |

There is no published security SLA, bounty program or guaranteed response time.

## Reporting a vulnerability privately

**Do not open a public issue containing exploit details, private keys, bypasses or sensitive logs.**

Use [GitHub private vulnerability reporting](https://github.com/20player11/kyrelis-blockchain/security/advisories/new) **if the repository has this feature enabled**. If that form is unavailable, contact the repository maintainer through an existing GitHub profile/contact channel and request a private reporting method without sending exploit details publicly. The feature may require the owner to enable it under **Settings → Security → Code security and analysis**.

Include a concise summary, affected revision/component, reproduction steps using disposable data, impact, and a suggested fix where possible. Redact secrets and personal information.

## Areas of particular interest

- Cosmos SDK/CometBFT consensus and transaction validity
- Signature verification, Keplr signing and broadcast paths
- `ukyrl` fee handling, chain ID, account sequence and replay resistance
- Genesis supply, staking and privilege boundaries
- Hub authentication, sessions, CSRF, reverse proxy and origin checks
- Operator authorization, local service isolation and command execution
- Key storage, backup, dependency and workflow supply-chain security

## Coordinated disclosure

The maintainer may request a private proof of concept, confirm the affected scope, prepare a patch and coordinate an advisory. Publication timing depends on severity and readiness; no fixed timeline is promised. Avoid testing against networks or accounts you do not own or have authorization to assess.

## Operator safety

Never commit `.local/`, keyring files, validator keys, `.env`, database files or production configurations. Keep RPC/admin endpoints restricted; publish only endpoints intentionally intended for public access. Use disposable testnet funds and separate keys for any internet-facing deployment.
