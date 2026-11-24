import {TxRaw} from 'cosmjs-types/cosmos/tx/v1beta1/tx';
import {SigningStargateClient} from '@cosmjs/stargate';
import {fromBech32} from '@cosmjs/encoding';
const $=id=>document.getElementById(id), chain=import.meta.env.VITE_CHAIN_ID, denom='ukyrl';
const coin={coinDenom:'KYRL',coinMinimalDenom:denom,coinDecimals:6};
let client,addr,csrf,busy=false;
async function api(route,payload){
 const res=await fetch('/api/'+route,payload===undefined?{}:{method:'POST',headers:{'Content-Type':'application/json','X-CSRFToken':csrf},body:JSON.stringify(payload)});
 if(!res.ok)throw Error('Požadavek selhal ('+res.status+'). Obnov stránku a přihlas se znovu.');
 const data=await res.json();if(data.csrf)csrf=data.csrf;return data;
}
function units(id){const s=$(id).value.trim().replace(',','.');if(!/^\d{1,8}(\.\d{1,6})?$/.test(s))throw Error('Zadej kladnou částku, nejvýše šest desetinných míst.');const [i,f='']=s.split('.');const n=BigInt(i)*1000000n+BigInt(f.padEnd(6,'0'));if(n<=0n)throw Error('Částka musí být kladná.');return n.toString();}
async function refresh(){const b=BigInt((await client.getBalance(addr,denom)).amount);$('balance').textContent=(b/1000000n)+','+(b%1000000n).toString().padStart(6,'0')+' KYRL';}
async function connect(){
 if(!window.keplr)throw Error('Nejdřív vytvoř peněženku v rozšíření Keplr a obnov tuto stránku.');
 await api('session');const prefix='kyr';
 await keplr.experimentalSuggestChain({chainId:chain,chainName:chain,rpc:location.origin+'/rpc',rest:location.origin+'/rest',bip44:{coinType:118},bech32Config:{bech32PrefixAccAddr:prefix,bech32PrefixAccPub:'kyrpub',bech32PrefixValAddr:'kyrvaloper',bech32PrefixValPub:'kyrvaloperpub',bech32PrefixConsAddr:'kyrvalcons',bech32PrefixConsPub:'kyrvalconspub'},currencies:[coin],stakeCurrency:coin,feeCurrencies:[{...coin,gasPriceStep:{low:.001,average:.002,high:.004}}]});
 await keplr.enable(chain);const signer=keplr.getOfflineSigner(chain);const [a]=await signer.getAccounts();addr=a.address;
 const c=await api('challenge',{});await api('login',{address:addr,message:c.message,signature:await keplr.signArbitrary(chain,addr,c.message)});
 client=await SigningStargateClient.connectWithSigner(location.origin+'/rpc',signer);
 if(await client.getChainId()!==chain)throw Error('RPC hlásí jiné ID sítě.');
 $('address').textContent=addr;$('wallet').hidden=false;$('join').hidden=true;await refresh();
 const response=await fetch('/rest/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED&pagination.limit=200');if(!response.ok)throw Error('Peněženka připojena. Seznam validátorů se nepodařilo načíst.');
 const data=await response.json();$('validators').replaceChildren(...data.validators.map(v=>new Option(v.description.moniker,v.operator_address)));
}
async function tx(action){
 const v=$('validators').value;let typeUrl,value;
 if(action==='send'){const to=$('recipient').value.trim();const parsed=fromBech32(to);if(parsed.prefix!=='kyr'||parsed.data.length!==20)throw Error('Neplatná adresa příjemce.');typeUrl='/cosmos.bank.v1beta1.MsgSend';value={fromAddress:addr,toAddress:to,amount:[{denom,amount:units('amount')}]};}
 else{if(!v)throw Error('Vyber validátora.');value={delegatorAddress:addr,validatorAddress:v};if(action==='claim')typeUrl='/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward';else{typeUrl='/cosmos.staking.v1beta1.'+(action==='delegate'?'MsgDelegate':'MsgUndelegate');value.amount={denom,amount:units('stake')};}}
 const msg={typeUrl,value};const gas=350000;const fee={gas:String(gas),amount:[{denom,amount:String(Math.max(100,Math.ceil(gas*.002)))}]};
 const pendingKey=chain+'/'+addr+'/pending';
 const previous=sessionStorage.getItem(pendingKey);
 if(previous){
   const receipt=await client.getTx(previous);
   if(!receipt)throw Error('Předchozí transakce zatím nemá potvrzení. Hash: '+previous);
   sessionStorage.removeItem(pendingKey);
   throw Error('Předchozí transakce: '+previous+' | blok '+receipt.height+' | kód '+receipt.code+'. Zkontroluj zůstatek.');
 }
 const response=await fetch('/rest/cosmos/auth/v1beta1/accounts/'+addr,{cache:'no-store'});
 if(!response.ok)throw Error('Účet nelze načíst. Ověř, že už má na síti prostředky.');
 const {account}=await response.json();
 if(account?.address!==addr)throw Error('Neočekávaný typ účtu.');
 const uint64=value=>{
   if(typeof value!=='string'||!/^\d+$/.test(value)||BigInt(value)>18446744073709551615n)throw Error('Neplatné číslo účtu nebo pořadí transakce.');
   return BigInt(value);
 };
 const signed=await client.sign(addr,[msg],fee,'',{chainId:chain,accountNumber:uint64(account.account_number),sequence:uint64(account.sequence)});
 const bytes=TxRaw.encode(signed).finish();
 const digest=await crypto.subtle.digest('SHA-256',bytes);
 const hash=Array.from(new Uint8Array(digest),b=>b.toString(16).padStart(2,'0')).join('').toUpperCase();
 sessionStorage.setItem(pendingKey,hash);
 let r;
 try {r=await client.broadcastTx(bytes);}
 catch(error){
   if(error.name==='BroadcastTxError')sessionStorage.removeItem(pendingKey);
   throw Error(error.message+' | Hash: '+hash);
 }
 sessionStorage.removeItem(pendingKey);if(r.code)throw Error('Transakce selhala: '+r.rawLog);
 $('status').textContent='Potvrzeno v bloku '+r.height+' · '+r.transactionHash;await refresh();
}
function bind(id,fn,event='click'){$(id).addEventListener(event,async e=>{e.preventDefault();if(busy)return;busy=true;try{$('status').textContent='';await fn();}catch(err){$('status').textContent=err.message;}finally{busy=false;}});}
bind('join',connect);bind('refresh',refresh);bind('exit',async()=>{await api('logout',{});location.reload()});bind('send',()=>tx('send'),'submit');for(const id of ['delegate','undelegate','claim'])bind(id,()=>tx(id));window.addEventListener('keplr_keystorechange',()=>location.reload());

