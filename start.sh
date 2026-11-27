#!/bin/sh
url=https://blockchain.tail3fb8f3.ts.net:8443/
case "$(uname -s)" in
  Darwin) exec open "$url" ;;
esac
if command -v xdg-open >/dev/null 2>&1 && [ -n "${DISPLAY:-}${WAYLAND_DISPLAY:-}" ]; then exec xdg-open "$url"; fi
printf "%s\n" "$url"
