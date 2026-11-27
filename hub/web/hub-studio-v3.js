const el=id=>document.getElementById(id);
function mode(create){
  el('signin').setAttribute('aria-pressed',String(!create));el('signup').setAttribute('aria-pressed',String(create));
  el('accessTitle').textContent=create?'Vytvoř si účet.':'Vítej zpátky.';
  el('instructions').textContent=create?'V Keplru si vytvoř peněženku. Pak ji připoj a prvním podpisem vytvoř svůj účet v Hubu.':'Připoj svou peněženku a potvrď přihlášení podpisem v Keplru.';
  el('get').hidden=!create;el('join').textContent=create?'Připojit novou peněženku →':'Připojit Keplr →';
}
mode(!window.keplr);el('signin').onclick=()=>mode(false);el('signup').onclick=()=>mode(true);
async function network(){try{const r=await fetch('/rpc/status',{cache:'no-store',signal:AbortSignal.timeout(8000)});const s=(await r.json()).result.sync_info;el('height').textContent=BigInt(s.latest_block_height).toLocaleString('cs');el('sync').textContent=s.catching_up?'Synchronizuje se':'Synchronizováno';el('time').textContent=new Date(s.latest_block_time).toLocaleTimeString('cs')}catch{for(const id of ['height','sync','time'])el(id).textContent='Nedostupné'}}
let timer;function screen(){const open=!el('wallet').hidden;el('welcome').hidden=open;clearInterval(timer);if(open){network();timer=setInterval(()=>{if(!document.hidden)network()},30000)}}new MutationObserver(screen).observe(el('wallet'),{attributes:true,attributeFilter:['hidden']});screen();
const reduce=matchMedia('(prefers-reduced-motion: reduce)');
const canvases=[...document.querySelectorAll('canvas')];let last=-100;
function draw(t){if(t-last>=65||(reduce.matches||document.body.dataset.motion==="off")){last=t;for(const c of canvases){const w=c.clientWidth,h=c.clientHeight;if(!w||!h)continue;const d=Math.min(devicePixelRatio,2);if(c.width!==Math.round(w*d)||c.height!==Math.round(h*d)){c.width=Math.round(w*d);c.height=Math.round(h*d)}const g=c.getContext('2d');g.setTransform(d,0,0,d,0,0);g.clearRect(0,0,w,h);g.lineWidth=1;const phase=(reduce.matches||document.body.dataset.motion==="off")?0:t/8000;for(let j=-14;j<44;j++){g.beginPath();for(let x=0;x<=w+5;x+=5){const u=x/w;const y=h*(.47+j*.035+.13*Math.sin(u*6+j*.12+phase)+.11*Math.cos(u*4-j*.08)-.22*u);if(x===0)g.moveTo(x,y);else g.lineTo(x,y)}g.strokeStyle='rgba(224,194,255,.26)';g.stroke()}}}if(!(reduce.matches||document.body.dataset.motion==="off"))frame=requestAnimationFrame(draw)}
let frame;function restart(){cancelAnimationFrame(frame);last=-100;draw(performance.now())}window.addEventListener('motionchange',restart);reduce.addEventListener('change',restart);new ResizeObserver(restart).observe(document.body);restart();

const $=id=>document.getElementById(id);
const db=import.meta.env.VITE_CHAIN_ID+'/desk/';
function read(key,def){try{return JSON.parse(localStorage.getItem(db+key))??def}catch{return def}}
function save(key,val){localStorage.setItem(db+key,JSON.stringify(val))}
function say(msg){$('feedback').textContent=msg}
const live=()=>$('address').textContent.trim();
let profile='',tab='home',accounts=read('accounts',{}),generation=0;
function button(label,act){const b=document.createElement('button');b.type='button';b.textContent=label;b.onclick=act;return b}
function label(a){return accounts[a]||a.slice(0,9)+'…'+a.slice(-6)}
function valid(a){const chars='qpzry9x8gf2tvdw0s3jn54khce6mua7l';if(!/^kyr1[0-9a-z]{38}$/.test(a))return false;const values=[3,3,3,0,11,25,18,...Array.from(a.slice(4),c=>chars.indexOf(c))];let sum=1;for(const n of values){if(n<0)return false;const top=sum>>>25;sum=((sum&33554431)<<5)^n;[0x3b6a57b2,0x26508e6d,0x1ea119fa,0x3d4233dd,0x2a1462b3].forEach((g,j)=>{if(top>>j&1)sum^=g})}return sum===1}
function render(){let h=document.getElementById('accountHistory');if(!h){h=document.createElement('a');h.id='accountHistory';h.textContent='Historie účtu';document.getElementById('quick').append(h)}h.href='/explorer.html?q='+encodeURIComponent(profile||live());const addresses=[...new Set([live(),...Object.keys(accounts)])].filter(Boolean);$('profile').replaceChildren(...addresses.map(a=>new Option(label(a)+(a===live()?' · připojeno':''),a)));$('profile').value=profile;
$('contacts').replaceChildren(...Object.entries(accounts).map(([a,name])=>{const row=document.createElement('div');row.className='contact';const title=document.createElement('strong');title.textContent=name;const addr=document.createElement('code');addr.textContent=a;row.append(title,addr,button('Použít pro převod',()=>{profile=live();tab='send';render();$('recipient').value=a;$('amount').focus()}),button('Odebrat',()=>{delete accounts[a];save('accounts',accounts);if(profile===a)profile=live();render()}));return row}));
for(const id of ['accounts','lookup','prefs'])$(id).hidden=tab!==id;
const titles={home:'Přehled účtu',send:'Odeslat KYRL',stake:'Staking',accounts:'Adresář peněženek',lookup:'Transakce',prefs:'Nastavení'};
$('pageTitle').textContent=titles[tab];
$('wallet').querySelector('.grid').hidden=!['home','send','stake'].includes(tab)||profile!==live();
document.querySelectorAll('[data-view]').forEach(panel=>panel.hidden=panel.dataset.view!==tab);$('observed').hidden=tab!=='home'||profile===live();
document.querySelectorAll('[data-tab]').forEach(b=>b.setAttribute('aria-pressed',String(b.dataset.tab===tab||(b.dataset.tab==="home"&&["send","stake"].includes(tab)))));
if(!$('observed').hidden)watch();}
async function watch(){const token=++generation,a=profile;$('watchAddress').textContent=a;$('watchBalance').textContent='Načítání…';try{const r=await fetch('/rest/cosmos/bank/v1beta1/balances/'+a+'/by_denom?denom=ukyrl',{signal:AbortSignal.timeout(8000)});if(!r.ok)throw Error();const n=BigInt((await r.json()).balance.amount);if(token===generation&&a===profile)$('watchBalance').textContent=(n/1000000n).toLocaleString('cs')+','+(n%1000000n).toString().padStart(6,'0')+' KYRL'}catch{if(token===generation)$('watchBalance').textContent='Zůstatek nedostupný'}}
document.querySelectorAll('[data-tab]').forEach(b=>b.onclick=()=>{tab=b.dataset.tab;say('');render()});
$('profile').onchange=e=>{profile=e.target.value;tab='home';render()};
$('activate').onclick=async()=>{try{const key=await window.keplr.getKey(import.meta.env.VITE_CHAIN_ID);if(key.bech32Address!==profile){sessionStorage.setItem(db+'switch',profile);say('Otevři Keplr, vyber účet '+label(profile)+' a potvrď přihlášení v Hubu.');return}sessionStorage.setItem(db+'switch',profile);location.reload()}catch{say('Odemkni Keplr a zkus účet připojit znovu.')}};
$('saveContact').onsubmit=e=>{e.preventDefault();const a=$('contactAddr').value.trim(),name=$('alias').value.trim();if(!valid(a)||!name)return say('Zkontroluj název a adresu kyr1 včetně kontrolního součtu.');accounts[a]=name;save('accounts',accounts);e.target.reset();render();say('Adresa uložena.')};
$('copyAddr').onclick=async()=>{try{await navigator.clipboard.writeText(profile);say('Adresa zkopírována.')}catch{say(profile)}};
$('findTx').onsubmit=async e=>{e.preventDefault();const hash=$('hash').value.trim().toUpperCase();$('receipt').textContent='Vyhledávám…';try{const r=await fetch('/rpc/tx?hash=0x'+hash,{signal:AbortSignal.timeout(8000)});const j=await r.json();if($('hash').value.trim().toUpperCase()!==hash)return;if(!r.ok||!j.result)throw Error();const t=j.result;$('receipt').textContent='Blok '+t.height+' · '+(Number(t.tx_result.code||0)===0?'Potvrzeno':'Selhalo · kód '+t.tx_result.code)}catch{$('receipt').textContent='Potvrzení se nepodařilo dohledat.'}};
function settings(){document.body.dataset.density=$('density').value;document.body.dataset.motion=$('motion').value;document.body.classList.toggle('private',read('privacy',false));$('privacy').textContent=read('privacy',false)?'Zobrazit zůstatky':'Skrýt zůstatky';window.dispatchEvent(new Event('motionchange'))}
for(const id of ['density','motion']){$(id).value=read(id,$(id).value);$(id).onchange=()=>{save(id,$(id).value);settings()}}$('privacy').onclick=()=>{save('privacy',!read('privacy',false));settings()};settings();
const draft=document.createElement('div');draft.className='actions draft';$('send').after(draft);draft.append(button('Uložit koncept',()=>{save('draft/'+live(),[$('recipient').value,$('amount').value]);say('Koncept uložen pro tento účet.')}),button('Načíst koncept',()=>{const d=read('draft/'+live(),null);if(!d)return say('Tento účet nemá uložený koncept.');[$('recipient').value,$('amount').value]=d;say('Koncept načten.')}),button('Vymazat formulář',()=>{$('send').reset();localStorage.removeItem(db+'draft/'+live());say('Formulář a koncept vymazány.')}));
function opened(){if($('wallet').hidden||!live())return;$('desk').hidden=false;profile=live();if(!accounts[profile]){accounts[profile]=label(profile);save('accounts',accounts)}render()}
new MutationObserver(opened).observe($('wallet'),{attributes:true,attributeFilter:['hidden']});opened();

async function resumeSwitch(){
 const target=sessionStorage.getItem(db+'switch');
 if(!target||!window.keplr)return;
 try{const key=await window.keplr.getKey(import.meta.env.VITE_CHAIN_ID);if(key.bech32Address===target){sessionStorage.removeItem(db+'switch');el('join').click()}}catch{}
}
resumeSwitch();

const systemLight = matchMedia('(prefers-color-scheme: light)');
const themeControl = document.getElementById('theme');
themeControl.value = read('theme', 'auto');
function applyTheme(){
  const light = themeControl.value === 'light' || (themeControl.value === 'auto' && systemLight.matches);
  document.documentElement.dataset.theme = light ? 'light' : 'dark';
  document.querySelector('meta[name="theme-color"]').content = light ? '#f6f5f9' : '#19181d';
}
systemLight.addEventListener('change',applyTheme);
themeControl.onchange = () => {applyTheme(); save('theme',themeControl.value)};
applyTheme();
function openTask(page, focusId){
  tab=page;
  render();
  $(focusId).focus();
}
$('gotoSend').onclick=()=>openTask('send','recipient');
$('gotoStake').onclick=()=>openTask('stake','validators');
document.querySelectorAll('[data-back]').forEach(b=>b.onclick=()=>openTask('home',b.closest('[data-view]').dataset.view==='send'?'gotoSend':'gotoStake'));
$('receive').onclick = () => {
  $('receiveAddress').textContent = live();
  $('copyResult').textContent = '';
  $('receiveDialog').showModal();
};
$('receiveCopy').onclick = async () => {
  try { await navigator.clipboard.writeText(live()); $('copyResult').textContent = 'Adresa zkopírována.'; }
  catch { $('copyResult').textContent = 'Označ a zkopíruj adresu výše.'; }
};

function updateQr(){const amount=$('requestedAmount').value.trim().replace(',','.');const img=$('paymentQR');img.hidden=true;if(amount&&(!/^\d{1,8}(\.\d{1,6})?$/.test(amount)||Number(amount)===0)){ $('copyResult').textContent='Zadej kladnou částku s nejvýše šesti desetinnými místy.';return} $('copyResult').textContent='';img.onload=()=>img.hidden=false;img.onerror=()=>$('copyResult').textContent='QR není dostupný. Zkopíruj adresu.';img.src='/api/payment-qr?'+new URLSearchParams({address:live(),amount})}
$('requestedAmount').oninput=updateQr;$('receive').addEventListener('click',updateQr);
