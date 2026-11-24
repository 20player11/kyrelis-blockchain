# This creates REAL Cosmos SDK source in WSL. Never silently deletes existing chain files.
set -euo pipefail
if ! command -v go >/dev/null || ! command -v ignite >/dev/null; then
  echo 'ERROR: Go and Ignite CLI v30 must first be installed inside WSL.' >&2
  exit 2
fi
if ! ignite version 2>&1 | grep -Eiq '30\.|v30'; then
  echo 'ERROR: Ignite CLI v30 is required.' >&2; exit 2
fi
mkdir -p "$HOME/KyrelisNetwork"
cd "$HOME/KyrelisNetwork"
if test -d kyrelis; then
  echo 'ERROR: Existing chain directory. It will NEVER be overwritten automatically.' >&2
  exit 2
fi
echo '[1/3] Generate real Cosmos SDK chain source'
ignite cosmos scaffold chain kyrelis --address-prefix kyr --default-denom ukyrl --no-module
cd kyrelis
echo '[2/3] Apply Kyrelis development genesis config'
# Replaced by the operator's bundled shell/config payload.
# GUI script appends the config using a verified here-document.
echo '[3/3] Build the Cosmos SDK chain (Go modules require internet)'
ignite cosmos chain build
printf '\nGenerated real chain: %s\n' "$HOME/KyrelisNetwork/kyrelis"
printf 'Not mainnet-ready: hard supply cap, fee floor, inflation policy and validator hardening still require code and tests.\n'
