import {fromBech32} from '@cosmjs/encoding';
const el=id=>document.getElementById(id);
const root=el('result'), bank='/rest/cosmos/bank/v1beta1/', staking='/rest/cosmos/staking/v1beta1/', tx='/rest/cosmos/tx/v1beta1/txs';
async function json(url){const r=await fetch(url,{signal:AbortSignal.timeout(10000),cache:'no-store'});if(!r.ok)throw Error('HTTP '+r.status);const j=await r.json();if(j.error)throw Error(j.error.message);return j}
function node(tag,text){const n=document.createElement(tag);n.textContent=text;return n}
function link(value,title=value){const a=node('a',title);a.href='?q='+encodeURIComponent(value);return a}
function amount(coins=[]){return coins.map(c=>c.denom==='ukyrl'?format(BigInt(c.amount))+' KYRL':c.amount+' '+c.denom).join(', ')||'0 KYRL'}
function format(n){return (n/1000000n).toLocaleString('cs')+','+(n%1000000n).toString().padStart(6,'0')}
function date(t){return new Date(t).toLocaleString('cs')}
async function history(address){
 const balance=await json(bank+'balances/'+address+'/by_denom?denom=ukyrl');
 root.append(node('h2','Účet'),node('code',address),node('p','Zůstatek '+amount([balance.balance])),node('h3','Platby'));
 const filter=document.createElement('select');filter.setAttribute('aria-label','Směr platby');filter.add(new Option('Odchozí','transfer.sender'));filter.add(new Option('Příchozí','transfer.recipient'));const list=document.createElement('div'),more=node('button','Další platby');root.append(filter,list,more);let page=1,ticket=0;
 async function load(reset){const id=++ticket;if(reset){page=1;list.replaceChildren()}more.disabled=true;more.hidden=false;
 try{const q=new URLSearchParams({query:filter.value+"='"+address+"'",page:String(page),limit:'20',order_by:'ORDER_BY_DESC'});
 const j=await json(tx+'?'+q);if(id!==ticket)return;
 const rows=j.tx_responses;if(!Array.isArray(rows))throw Error('Neplatná odpověď indexu');
 for(const r of rows){const line=document.createElement('p');line.append(link(r.txhash,r.txhash.slice(0,14)+'…'),document.createTextNode(' · '+date(r.timestamp)+' · '+(r.code?'Selhalo':'Potvrzeno')+' · poplatek '+amount(r.tx.auth_info.fee.amount)));
 for(const m of r.tx.body.messages){if(m['@type']==='/cosmos.bank.v1beta1.MsgSend'&&(m.to_address===address||m.from_address===address))line.append(node('br',''),document.createTextNode(amount(m.amount)+' → '),link(m.to_address))}
 list.append(line)}if(!rows.length&&page===1)list.append(node('p','Pro tento směr nejsou v indexu žádné platby.'));more.hidden=rows.length<20;page++}
 catch(e){if(id===ticket)list.append(node('p','Historii se nepodařilo načíst: '+e.message))}finally{if(id===ticket)more.disabled=false}}
 filter.onchange=()=>load(true);more.onclick=()=>load(false);await load(true)
}
async function detail(q){
 if(/^\d{1,18}$/.test(q)){const b=(await json('/rpc/block?height='+q)).result.block;root.append(node('h2','Blok '+b.header.height),node('p',date(b.header.time)+' · '+(b.data.txs||[]).length+' transakcí'));for(const raw of b.data.txs||[]){const digest=await crypto.subtle.digest('SHA-256',Uint8Array.from(atob(raw),s=>s.charCodeAt(0)));root.append(link(Array.from(new Uint8Array(digest),b=>b.toString(16).padStart(2,'0')).join('').toUpperCase()),node('br',''))}return}
 if(/^[\da-f]{64}$/i.test(q)){const r=(await json(tx+'/'+q.toUpperCase())).tx_response;root.append(node('h2','Transakce'),node('code',r.txhash),node('p',date(r.timestamp)+' · '+(r.code?'Selhalo · '+r.code:'Potvrzeno')),link(r.height,'Blok '+r.height),node('p','Poplatek '+amount(r.tx.auth_info.fee.amount)));for(const m of r.tx.body.messages){const pre=node('pre',JSON.stringify(m,null,2));pre.style.cssText='white-space:pre-wrap;overflow-wrap:anywhere';root.append(pre)}return}
 const decoded=fromBech32(q);if(decoded.prefix==='kyrvaloper'){const v=(await json(staking+'validators/'+q)).validator;root.append(node('h2',v.description.moniker),node('code',q),node('p','Podíl validátoru na odměnách: '+(Number(v.commission.commission_rates.rate)*100).toLocaleString('cs')+' %'),node('p','Vázáno: '+format(BigInt(v.tokens))+' KYRL'),node('p',v.jailed?'Validátor je vyřazen':v.status));return}if(decoded.prefix!=='kyr')throw Error('Neplatný prefix adresy');await history(q)}
el('search').onsubmit=e=>{e.preventDefault();location.search=new URLSearchParams({q:el('query').value.trim()})};
const query=new URLSearchParams(location.search).get('q');if(query){el('query').value=query;detail(query).catch(e=>root.append(node('p','Dotaz se nepodařilo dokončit: '+e.message)))}else root.hidden=true;
async function health(){try{const s=(await json('/rpc/status')).result;const age=Math.max(0,Math.round((Date.now()-Date.parse(s.sync_info.latest_block_time))/1000));el('health').replaceChildren(document.createTextNode(s.node_info.network+' · '),link(s.sync_info.latest_block_height,'Blok '+s.sync_info.latest_block_height),document.createTextNode(' · poslední blok před '+age+' s'+(age>60?' · Zpoždění bloků':'')))}catch(e){el('health').textContent='RPC nedostupné · '+e.message}}health();setInterval(()=>{if(!document.hidden)health()},15000);
json(staking+'validators?status=BOND_STATUS_BONDED&pagination.limit=200').then(j=>{for(const v of j.validators)el('validators').append(link(v.operator_address,v.description.moniker),node('br',''))}).catch(e=>el('validators').textContent=e.message);
el('diagnose').onclick=async()=>{el('diagnose').disabled=true;const log=s=>el('report').append(document.createTextNode(s+'\n'));el('report').replaceChildren();log('Keplr: '+(window.keplr?'nalezen':'nedetekován'));for(const [name,url] of [['RPC','/rpc/status'],['REST',bank+'supply/by_denom?denom=ukyrl']]){const t=performance.now();try{await json(url);log(name+': OK · '+Math.round(performance.now()-t)+' ms')}catch(e){log(name+': '+e.message)}}
 try{await new Promise((ok,fail)=>{const url=new URL('/rpc/websocket',location.href);url.protocol=location.protocol==='https:'?'wss:':'ws:';const ws=new WebSocket(url);const t=setTimeout(()=>{ws.close();fail(Error('timeout'))},7000);ws.onopen=()=>ws.send(JSON.stringify({jsonrpc:'2.0',id:1,method:'status',params:{}}));ws.onmessage=e=>{try{if(!JSON.parse(e.data).result.node_info.network)throw Error();ok()}catch{fail(Error('neplatná odpověď'))}finally{clearTimeout(t);ws.close()}};ws.onerror=ws.onclose=()=>{clearTimeout(t);fail(Error('spojení zavřeno'))}});log('WebSocket: OK')}catch(e){log('WebSocket: '+e.message)}log('Podpis a broadcast se ověřují při skutečné transakci v Hubu.');el('diagnose').disabled=false};
