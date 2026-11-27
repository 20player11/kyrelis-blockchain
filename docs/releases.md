# Vydání uzlu a Hubu

Chain ID označuje konkrétní síť. Verze softwaru označuje zdrojový commit. Mainnet může během života používat více verzí uzlu se stejným chain ID. Stav větve main sám o sobě nepotvrzuje nasazení na Pi.

## Příprava změny

Změnu vyvíjej ve větvi a popiš ji v CHANGELOG.md pod Nevydané. U změny pravidel konsenzu popiš dopad na existující stav a požadavky na koordinaci validátorů. Pro vývojovou zkoušku jsou pod tools/development/ oddělené spouštěče uzlu, Hubu a konzole. Tyto nástroje pracují s místním prostředím.

Sestavení uzlu spouštěj v chain/ s verzí Go uvedenou v go.mod:

    go test ./app -run TestNativeRules
    go build -trimpath -o kyrelisd ./cmd/kyrelisd

Frontend lze sestavit v hub/web/ příkazy npm ci a npm run build. Pro síťový build nastav VITE_CHAIN_ID na cílové ID. Konzoli ověř příkazem go test ./... v operator/.

## Před zveřejněním

1. Ověř změny na devnetu. Zkontroluj růst výšky, připojení Keplru a potvrzený testovací převod. Zaznamenej hash transakce a commit.
2. V pull requestu uveď kompatibilní chain ID, postup aktualizace a výsledek kontroly obnovy. Změny genesis a změny binárky posuzuj odděleně.
3. Po schválení označ přesný commit tagem vMAJOR.MINOR.PATCH. Do GitHub Release přidej podporované architektury, SHA-256 artefaktů, požadavky a případnou aktivační výšku.
4. Po nasazení doplň do poznámek skutečně ověřenou síť, výšku bloku a verzi. U neúspěšné migrace nejprve prověř kompatibilitu databáze se starší binárkou.

Veřejná konfigurace sítí je v networks/. Změny endpointů promítni také do dokumentace a příslušného Keplr registry PR. Genesis, její ověřený otisk a identitu peerů publikuj teprve po kontrole proti provozovanému uzlu. Starý otisk kyrelis-1 nelze použít pro kyrelis-2.
