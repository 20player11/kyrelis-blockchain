# Main branch protection (KYRELIS)

## Current state

This repository has a reviewed JSON **ruleset proposal** in `docs/github-rulesets/kyrelis-main.json`. **A committed JSON file does not activate protection.** The ruleset must be imported and saved in GitHub repository settings by the owner.

## Import on GitHub

1. Open [Repository Rulesets](https://github.com/20player11/kyrelis-blockchain/settings/rules).
2. Choose **New ruleset → Import a ruleset**, or create a branch ruleset and paste/import the contents of [kyrelis-main.json](github-rulesets/kyrelis-main.json).
3. Target: **Branch**, default branch (`main`). Verify enforcement is **Active**.
4. Verify rules: **restrict deletions**, **block force pushes**, **require pull request before merging**, and **require conversation resolution**.
5. Set **required approving reviews = 0** for the current solo-maintainer workflow. Do not enable code-owner review yet: `@20player11` cannot independently approve their own PR. Enable mandatory 1+ independent review once a trusted second maintainer is available.
6. Keep the **bypass list empty** for strict protection. If GitHub warns that a ruleset cannot be enforced under the repository plan, inspect the warning rather than assuming protection is active.
7. Save and verify the ruleset appears as **Active**, and that direct edits to `main` are blocked.

## After CI is verified

When [Quality checks](https://github.com/20player11/kyrelis-blockchain/actions) have run successfully on pull requests, consider requiring these exact check names:

- `Python syntax`
- `Operator Go tests`
- `Hub frontend build`

**Do not require checks blindly**: if workflows never run or have different names, merges may be blocked. The CI workflow currently does not execute the full Cosmos SDK tests.

## Additional protection

- In **Settings → Actions**, give workflows the minimum read-only permissions required.
- Enable **Dependabot alerts** and **Private vulnerability reporting** under repository security settings when supported.
- Once another trusted maintainer joins, enable **1 independent approving review**, stale review dismissal, and optionally mandatory code-owner review.
- Protect release tags with a **separate tag ruleset** if you start publishing signed releases.
- For production networks, require documented upgrades, deterministic chain testing, and explicit review of any change involving consensus, genesis, key handling, fees or wallet transaction signing.

## Verification checklist

- [ ] Ruleset is **Active** in GitHub Settings (not just committed as a file)
- [ ] Branch target matches `main`
- [ ] Force pushes are rejected
- [ ] Branch deletion is rejected
- [ ] Changes must arrive by PR
- [ ] Review conversations must be resolved
- [ ] CI required-check names have been verified before enabling them
