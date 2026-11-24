set -euo pipefail
cd "$HOME/KyrelisNetwork/kyrelis"
if ! test -f go.mod; then echo 'ERROR: Real Cosmos project not found. Use Create real chain first.' >&2; exit 2;fi
printf '%s\n' "$$" > "$HOME/.kyrelis-ignite-serve.pid"
echo 'Starting REAL CometBFT devnet via Ignite. Never use Ignite serve for mainnet.'
exec ignite cosmos chain serve
