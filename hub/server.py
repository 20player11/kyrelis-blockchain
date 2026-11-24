import base64, hashlib, json, os, secrets, time
from urllib.parse import urlsplit
from django.conf import settings
ORIGIN=os.environ["PUBLIC_ORIGIN"].rstrip("/")
host=urlsplit(ORIGIN)
local = os.environ.get("KYRELIS_WIN_CHILD") == "1" and ORIGIN == "http://127.0.0.1:5173"
if (host.scheme!="https" and not local) or not host.hostname or host.path: raise ValueError("PUBLIC_ORIGIN must be an HTTPS origin")
settings.configure(SECRET_KEY=os.environ["DJANGO_SECRET_KEY"],DEBUG=False,SECURE_PROXY_SSL_HEADER=("HTTP_X_FORWARDED_PROTO","https"),ALLOWED_HOSTS=[host.hostname],CSRF_TRUSTED_ORIGINS=[ORIGIN],ROOT_URLCONF=__name__,MIDDLEWARE=["django.middleware.security.SecurityMiddleware","django.contrib.sessions.middleware.SessionMiddleware","django.middleware.csrf.CsrfViewMiddleware","django.middleware.clickjacking.XFrameOptionsMiddleware"],INSTALLED_APPS=["django.contrib.sessions"],DATABASES={"default":{"ENGINE":"django.db.backends.sqlite3","NAME":os.environ.get("DATABASE","/var/lib/kyrelis-hub/accounts.sqlite3"),"OPTIONS":{"timeout":10}}},SESSION_COOKIE_SECURE=not local,CSRF_COOKIE_SECURE=not local,SESSION_COOKIE_SAMESITE="Strict",SESSION_COOKIE_AGE=86400,DATA_UPLOAD_MAX_MEMORY_SIZE=8192,USE_TZ=True)
import django
django.setup()
from django.core.wsgi import get_wsgi_application
from django.db import connection,transaction
from django.http import JsonResponse
from django.middleware.csrf import get_token,rotate_token
from django.urls import path
from django.views.decorators.http import require_GET,require_POST
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.asymmetric import ec,utils
from cryptography.exceptions import InvalidSignature
from bech32 import bech32_encode,convertbits
CHAIN=os.environ["KYRELIS_CHAIN_ID"]
def body(r):
    if r.content_type!="application/json" or len(r.body)>8192: raise ValueError()
    return json.loads(r.body)
def address(key):
    raw=hashlib.new("ripemd160",hashlib.sha256(key).digest()).digest()
    return bech32_encode("kyr",convertbits(raw,8,5))
def signed_bytes(addr,message):
    doc={"account_number":"0","chain_id":"","fee":{"amount":[],"gas":"0"},"memo":"","msgs":[{"type":"sign/MsgSignData","value":{"data":base64.b64encode(message.encode()).decode(),"signer":addr}}],"sequence":"0"}
    return json.dumps(doc,sort_keys=True,separators=(",",":")).encode()
@require_GET
def session(r):return JsonResponse({"csrf":get_token(r),"address":r.session.get("address")})
@require_POST
def challenge(r):
    if not r.session.session_key:
        r.session.create()
        r.session["pending_login"] = True
    nonce=secrets.token_urlsafe(32)
    message=f"KYRELIS login\n{ORIGIN}\n{CHAIN}\n{nonce}"
    with connection.cursor() as c:
        c.execute("DELETE FROM challenge WHERE expires < %s",[int(time.time())])
        c.execute("INSERT OR REPLACE INTO challenge VALUES (%s,%s,%s)",[r.session.session_key,message,int(time.time())+120])
    return JsonResponse({"message":message})
@require_POST
def login(r):
    try:
        data=body(r); sig=data["signature"]; pub=sig["pub_key"]
        if pub["type"]!="tendermint/PubKeySecp256k1":raise ValueError()
        key=base64.b64decode(pub["value"],validate=True); raw=base64.b64decode(sig["signature"],validate=True)
        if len(key)!=33 or len(raw)!=64 or not r.session.session_key:raise ValueError()
        addr=address(key)
        if addr!=data["address"]:raise ValueError()
        with transaction.atomic(),connection.cursor() as c:
            c.execute("DELETE FROM challenge WHERE sid=%s AND message=%s AND expires>=%s",[r.session.session_key,data["message"],int(time.time())])
            if c.rowcount!=1:raise ValueError()
            pk=ec.EllipticCurvePublicKey.from_encoded_point(ec.SECP256K1(),key)
            pk.verify(utils.encode_dss_signature(int.from_bytes(raw[:32],"big"),int.from_bytes(raw[32:],"big")),signed_bytes(addr,data["message"]),ec.ECDSA(hashes.SHA256()))
            c.execute("INSERT OR IGNORE INTO account VALUES (%s,%s)",[addr,int(time.time())])
        r.session.cycle_key();r.session["address"]=addr;rotate_token(r)
        return JsonResponse({"address":addr,"csrf":get_token(r)})
    except (ValueError,TypeError,KeyError,InvalidSignature):return JsonResponse({"error":"Podpis je neplatný nebo vypršel."},status=403)
@require_POST
def logout(r):
    r.session.flush();return JsonResponse({"csrf":get_token(r)})
def payment_qr(request):
    from bech32 import bech32_decode
    from django.http import HttpResponse
    from io import BytesIO
    import re
    import qrcode
    from qrcode.image.svg import SvgPathImage
    from urllib.parse import urlencode
    address = request.GET.get("address", "")
    amount = request.GET.get("amount", "")
    prefix, words = bech32_decode(address)
    if prefix != "kyr" or len(address) != 42 or not words:
        return JsonResponse({"error":"Invalid address"},status=400)
    if amount and (not re.fullmatch(r"[0-9]{1,8}(\.[0-9]{1,6})?",amount) or not int(amount.replace(".",""))):
        return JsonResponse({"error":"Invalid amount"},status=400)
    payload = address if not amount else "kyrelis:" + address + "?" + urlencode({"chain":CHAIN,"amount":amount})
    out = BytesIO()
    qrcode.make(payload,image_factory=SvgPathImage).save(out)
    return HttpResponse(out.getvalue(),content_type="image/svg+xml")
urlpatterns=[path("api/payment-qr", payment_qr)]+[path("api/"+n,f) for n,f in [("session",session),("challenge",challenge),("login",login),("logout",logout)]]
application=get_wsgi_application()
if __name__=="__main__":
    from django.core.management import call_command
    call_command("migrate",interactive=False)
    with connection.cursor() as c:
        c.execute("CREATE TABLE IF NOT EXISTS account(address TEXT PRIMARY KEY, created INTEGER NOT NULL)")
        c.execute("CREATE TABLE IF NOT EXISTS challenge(sid TEXT PRIMARY KEY,message TEXT NOT NULL,expires INTEGER NOT NULL)")
