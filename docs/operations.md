# Provoz KYRELIS na Raspberry Pi

Na produkčním Pi běží validátor jako uživatelská služba a webový Hub jako systémová služba. Následující příkazy spusť na Pi pod účtem, který instaloval uzel.

| Součást | Mainnet | Devnet |
|---|---|---|
| Validátor (systemctl --user) | kyrelis-mainnet | kyrelis-dev-validator |
| Web (systemctl) | kyrelis-main-hub | kyrelis-hub |
| Lokální RPC / REST | 27657 / 2317 | 26657 / 1317 |
| Web za HTTPS proxy | 28188 | 8188 |

## Kontrola po restartu

    systemctl --user is-active kyrelis-mainnet kyrelis-dev-validator
    systemctl is-active kyrelis-main-hub kyrelis-hub tailscaled
    loginctl show-user "$USER" -p Linger
    curl -fsS http://127.0.0.1:27657/status
    tailscale funnel status

Ověř rostoucí výšku dvakrát s odstupem několika sekund a pak otevři /rpc/status také na veřejném Hubu. Když lokální RPC funguje a veřejný Hub selže, prověř službu Hubu, port a Funnel. Výpis služby získáš přes journalctl --user -u kyrelis-mainnet -n 50 nebo sudo journalctl -u kyrelis-main-hub -n 50.

## Záloha a obnova

Kompletní instalátor umisťuje pod ~/.local/share/kyrelis-setup/ skripty backup-main.sh a backup-dev.sh. Během konzistentní zálohy se příslušný validátor krátce zastaví. Výsledný šifrovaný archiv přenes mimo Pi a ověř možnost obnovy. Samostatně zálohuj databáze Hubu, nastavení webhooků a jejich kurzory, stav indexeru a konfiguraci služeb. Archiv uzlu tyto součásti nepokrývá.

Při obnově zachovej genesis, databáze uzlu, klíč validátora a poslední stav podepisování z jedné konzistentní kopie. Před spuštěním ověř, že tentýž konsenzuální klíč nepoužívá druhý běžící validátor. Uzel s novou genesis patří do nové sítě. Seed uživatelské peněženky sám neobnoví bloky ani stav validátora.
