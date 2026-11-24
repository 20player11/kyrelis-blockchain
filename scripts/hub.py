"""Start the local web wallet alongside a running starter chain."""
import json
import os
from pathlib import Path
import secrets
import subprocess
import sys
from urllib.request import urlopen
from devnet import ROOT, HOME, STATE

def main():
    chain = json.loads((HOME / "config/genesis.json").read_text())["chain_id"]
    with urlopen("http://127.0.0.1:26657/status",timeout=5) as res:
        if json.load(res)["result"]["node_info"]["network"] != chain:
            sys.exit("RPC chain ID does not match the local genesis.")
    envdir=STATE/"hub-venv"
    python=envdir/("Scripts/python.exe" if os.name=="nt" else "bin/python")
    if not python.exists():
        subprocess.run([sys.executable,"-m","venv",str(envdir)],check=True)
    hub=ROOT/"hub"
    web=hub/"web"
    subprocess.run([str(python),"-m","pip","install","-r","requirements.txt"],cwd=hub,check=True)
    npm="npm.cmd" if os.name=="nt" else "npm"
    subprocess.run([npm,"ci"],cwd=web,check=True)
    env={**os.environ,"VITE_CHAIN_ID":chain,"KYRELIS_CHAIN_ID":chain,"KYRELIS_WIN_CHILD":"1",
         "PUBLIC_ORIGIN":"http://127.0.0.1:5173","DATABASE":str(STATE/"accounts.sqlite3"),
         "DJANGO_SECRET_KEY":secrets.token_urlsafe(48)}
    subprocess.run([npm,"run","build"],cwd=web,env=env,check=True)
    subprocess.run([str(python),"server.py"],cwd=hub,env=env,check=True)
    api=subprocess.Popen([str(python),"-m","waitress","--listen=127.0.0.1:18188","server:application"],cwd=hub,env=env)
    try:
        subprocess.run([npm,"exec","--","vite","preview"],cwd=web,env=env,check=True)
    finally:
        api.terminate()
        api.wait()

if __name__=="__main__":
    main()
