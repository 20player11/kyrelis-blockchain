"""Build and run an independent, single-validator KYRELIS development chain."""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess
from tempfile import TemporaryDirectory
from urllib.request import urlopen
from uuid import uuid4

ROOT = Path(__file__).resolve().parents[1]
STATE = ROOT / ".local"
HOME = STATE / "node"
BIN = ROOT / "bin" / ("kyrelisd.exe" if os.name == "nt" else "kyrelisd")


def build():
    BIN.parent.mkdir(exist_ok=True)
    subprocess.run(["go", "build", "-mod=readonly", "-trimpath", "-o", str(BIN),
                    "./cmd/kyrelisd"], cwd=ROOT / "chain", check=True,
                   env={**os.environ, "CGO_ENABLED": "0"})


def initialize(chain_id, moniker):
    if HOME.exists():
        actual = json.loads((HOME / "config/genesis.json").read_text())["chain_id"]
        if chain_id and actual != chain_id:
            raise SystemExit(f"Existing node uses {actual}. Use another checkout for a new chain.")
        return
    chain_id = chain_id or f"kyrelis-{uuid4().hex[:10]}-1"
    if not re.fullmatch(r"[a-z0-9][a-z0-9-]{0,48}", chain_id):
        raise SystemExit("Use 1–49 lowercase letters, digits or hyphens for the chain ID.")
    STATE.mkdir(exist_ok=True, mode=0o700)
    with TemporaryDirectory(dir=STATE) as staging:
        node = Path(staging) / "node"

        def run(*args):
            return subprocess.run([str(BIN), *args, "--home", str(node)], check=True,
                                  stdout=subprocess.PIPE, text=True, encoding="utf-8").stdout.strip()

        run("init", moniker, "--chain-id", chain_id)
        for name, units in [("validator", 1_000_000_000), ("treasury", 20_999_000_000_000)]:
            # The test keyring is for disposable local development only.
            subprocess.run([str(BIN),"keys","add",name,"--keyring-backend","test","--output","json",
                            "--home",str(node)], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, check=True)
            address = run("keys","show",name,"-a","--keyring-backend","test")
            run("genesis","add-genesis-account",address,f"{units}ukyrl")
        run("genesis","gentx","validator","100000000ukyrl","--chain-id",chain_id,"--keyring-backend","test")
        run("genesis","collect-gentxs")
        path = node / "config/genesis.json"
        genesis = json.loads(path.read_text())
        genesis["app_state"]["bank"]["denom_metadata"] = [dict(description="Kyrelis",base="ukyrl",
            display="KYRL",name="Kyrelis",symbol="KYRL",denom_units=[
                dict(denom="ukyrl",exponent=0,aliases=[]),dict(denom="KYRL",exponent=6,aliases=[])])]
        path.write_text(json.dumps(genesis,indent=2),encoding="utf-8")
        run("genesis","validate")
        node.rename(HOME)
    print("Created",chain_id)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action",choices=["build","init","start","status"],default="start",nargs="?")
    parser.add_argument("--chain-id")
    parser.add_argument("--moniker",default="local-validator")
    args = parser.parse_args()
    if args.action == "status":
        with urlopen("http://127.0.0.1:26657/status",timeout=5) as response:
            data = json.load(response)["result"]
        print(data["node_info"]["network"], "block",data["sync_info"]["latest_block_height"])
        return
    if args.action == "build" or not BIN.exists():
        build()
    if args.action != "build":
        initialize(args.chain_id,args.moniker)
    if args.action == "start":
        subprocess.run([str(BIN),"start","--home",str(HOME),"--minimum-gas-prices","0.001ukyrl",
                        "--p2p.laddr","tcp://127.0.0.1:26656","--rpc.laddr","tcp://127.0.0.1:26657",
                        "--api.enable","--api.address","tcp://127.0.0.1:1317"],check=True)


if __name__ == "__main__":
    main()
