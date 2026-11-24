# Collaboration and project governance

KYRELIS is an educational, evolving blockchain project. This document explains how to work together; it is not a claim of decentralization, production readiness, or formal audit.

## Collaboration principles

- Be respectful, specific, and evidence-driven.
- Prefer public issues for non-sensitive decisions and reproducible bug reports.
- Discuss complex features before starting large implementations.
- Keep changes reviewable and document decisions that affect users.
- Never post private credentials, seed phrases, security exploits or unreleased vulnerability details.

## From idea to implementation

1. **Propose:** Choose an issue form. State the problem, expected outcome, and alternatives.
2. **Scope:** Identify affected components and potential changes to chain data or wallet compatibility.
3. **Design:** For significant changes, write a brief proposal in the issue (motivation, specification, risks, test plan, migration).
4. **Implement:** Work on a feature branch. Include tests, documentation and a rollback strategy where meaningful.
5. **Review:** Address functional, security and cross-platform feedback.
6. **Merge/release:** Repository maintainers own final merge and release decisions; deployment is a separate controlled step.

## Review priorities

| Area | Minimum review focus |
| --- | --- |
| Cosmos SDK chain | Determinism, genesis compatibility, fees, consensus, staking, upgrade impacts |
| Wallet / Keplr | Signing, sequence, account lookup, RPC/REST, transport, broadcast, on-chain confirmation |
| Hub | Session integrity, CSRF, origin checks, data handling, dependencies |
| Operator | Local authorization, origin/host validation, subprocess safety, RPC handling |
| Launch scripts | Idempotence, cross-platform paths, safe file permissions, secrets and local state |
| Docs / UI | Clear instructions, accessibility, accurate examples, verified screenshots |

## Maintainers and ownership

The repository owner `@20player11` is the initial contact for code review and governance. `.github/CODEOWNERS` expresses default review ownership; GitHub's branch rules are configured separately. No promise of review turnaround or external support is made.

## Handling disagreements

Prefer technical evidence, small prototypes and documented trade-offs. Where consensus is not reached, maintainers make a final decision and explain relevant constraints. Follow the [Code of Conduct](CODE_OF_CONDUCT.md) in all project spaces.

## Security and releases

Do not publish exploitable reports before coordinated handling under [SECURITY.md](SECURITY.md). Development networks may reset; never rely on disposable keys or genesis state for permanent value.
