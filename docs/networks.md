# Síť a peněženka

| Rozhraní | Mainnet | Devnet |
|---|---|---|
| Chain ID | kyrelis-2 | kyrelis-devnet-3 |
| RPC | https://blockchain.tail3fb8f3.ts.net:8443/rpc | https://blockchain.tail3fb8f3.ts.net/rpc |
| REST / LCD | https://blockchain.tail3fb8f3.ts.net:8443/rest | https://blockchain.tail3fb8f3.ts.net/rest |

V Hubu připoj požadovaný účet Keplru a před podpisem zkontroluj síť, příjemce, částku a poplatek. Soukromé klíče zůstávají v peněžence. Převod má výsledek až po zařazení do bloku. Potvrzený výsledek s kódem 0 označuje úspěšné provedení.

## Ověření připojení

1. Připoj za adresu RPC cestu `/status`. Zkontroluj `result.node_info.network` a `result.sync_info.latest_block_height`. Při dalším načtení má výška růst.
2. Pro zůstatek připoj za REST cestu `/cosmos/bank/v1beta1/balances/ADRESA`. Nahraď ADRESA vlastní veřejnou adresou. Částku u `ukyrl` vyděl milionem. Prázdné pole balances znamená, že odpověď neobsahuje žádný zůstatek.
3. Při chybě „Failed to fetch“ v Keplru porovnej hostitele i port s tabulkou. Dřívější endpointy používaly hostitele raspberrypi.tail3fb8f3.ts.net. V konfiguraci Keplru patří do pole RPC základní adresa končící /rpc a do LCD adresa končící /rest.

Původní mainnet kyrelis-1 a devnet kyrelis-devnet-2 mají samostatnou historii. Současné sítě začaly novou genesis. Zápis chain ID v registru peněženek slouží k připojení účtu; aktuální stav zůstatků poskytuje příslušná síť.
