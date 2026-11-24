## Summary
<!-- What changed and why? Link related issue(s). -->

Closes #

## Change type
- [ ] Bug fix
- [ ] Feature
- [ ] Documentation
- [ ] Refactor / maintenance
- [ ] Security-sensitive change
- [ ] Consensus / economics / genesis / protocol change

## Components affected
- [ ] `chain/` (Cosmos SDK)
- [ ] `hub/` (Django + Vite / Keplr)
- [ ] `operator/` (Go)
- [ ] `scripts/` and launchers
- [ ] Docs / GitHub configuration

## Testing and evidence
<!-- Include exact commands, versions, results, screenshots or redacted logs. If untested, explain why. -->
- [ ] `cd operator && go test ./...`
- [ ] `cd chain && go test ./app ./cmd/kyrelisd/cmd` (where supported)
- [ ] `cd hub/web && npm ci && npm run build`
- [ ] Manual Windows / Linux smoke test where relevant
- [ ] Wallet transfer / delegation / transaction confirmation where relevant
- [ ] Documentation updated as appropriate

## Risk, compatibility and rollout
<!-- Include key handling, chain ID, RPC/REST, gas, genesis, migrations, backwards compatibility and rollback where applicable. -->
**Risk:** Low / Medium / High

**Migration / rollout:**

**Rollback plan:**

## Contributor checklist
- [ ] Scoped change; no unrelated modifications
- [ ] No seed phrases, private keys, test keyrings, secrets, personal data or generated node state committed
- [ ] No real assets used during tests
- [ ] New behavior has appropriate tests or an explanation
- [ ] `SECURITY.md` followed for vulnerabilities
- [ ] Existing licenses and third-party notices respected

> Never merge a consensus-critical change solely because CI is green. Obtain explicit protocol and security review.
