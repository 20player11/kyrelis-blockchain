set -euo pipefail
printf 'WSL: OK\n'
printf 'Linux: '; uname -srm
printf 'Go: '; go version
printf 'Ignite: '; ignite version
if ! ignite version 2>&1 | grep -Eiq '30\.|v30'; then
  echo 'ERROR: Ignite CLI v30 required.' >&2
  exit 2
fi
printf 'Dependencies: found\n'
