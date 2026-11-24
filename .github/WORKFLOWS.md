# GitHub repository automation and settings

Files in `.github/` provide community issue forms, a pull-request template, default review ownership, dependency update requests and CI checks.

## After merging the community setup

Repository maintainers should check these GitHub settings manually. A repository file **cannot** enable every platform-level feature:

1. **Settings → General → Features:** Enable **Issues** and, if desired, **Discussions**. Configure Discussion categories for Ideas and Q&A before relying on discussion templates.
2. **Settings → Security → Code security and analysis:** Enable **Private vulnerability reporting** and, where available, **Dependabot alerts** and security updates.
3. **Settings → Rules → Rulesets** (or branch protection): protect `main`, require pull requests and appropriate passing CI checks, and require review for sensitive changes. Set rules only after verifying that CI jobs complete on your plan/runner.
4. **Settings → Actions → General:** apply least-privilege workflow permissions; keep `GITHUB_TOKEN` read-only by default and avoid untrusted secrets on pull-request workflows.
5. **Settings → General → Pull Requests:** choose the desired merge strategy and enable automatically deleting head branches after merge if appropriate.

The included `.github/CODEOWNERS` lists `@20player11` as maintainer; review enforcement depends on repository settings and GitHub plan constraints.

## CI scope

`quality.yml` checks Python syntax, Go operator tests and the Hub frontend build. The Cosmos SDK chain needs an appropriate Go toolchain and additional integration tests before CI can claim comprehensive protocol coverage. Passing these checks does **not** certify the blockchain, wallet safety, public endpoints or consensus correctness.

## Adding more automation safely

Prefer pinned dependencies, scoped permissions, minimal new third-party actions and reliable checks over decorative badges or unnecessary workflows. Review Dependabot PRs individually; do not auto-merge chain dependency upgrades.
