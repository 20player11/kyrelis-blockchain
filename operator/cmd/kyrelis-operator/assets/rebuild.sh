set -euo pipefail
cd "$HOME/KyrelisNetwork/kyrelis"
if ! test -f go.mod; then
  echo 'ERROR: No generated real chain found. Use Generate Cosmos chain first.' >&2
  exit 2
fi
echo 'Rebuilding existing genuine Cosmos SDK source (no files removed)...'
ignite cosmos chain build
printf 'Real chain build completed. Not yet mainnet-ready.\n'
