const fs = require('node:fs');
const path = require('node:path');
const entry = require.resolve('@cosmjs/encoding');
const file = path.join(path.dirname(entry), 'bech32.js');
const original = fs.readFileSync(file,'utf8');
const fixed = original.replace(/\b(?:Number\.POSITIVE_INFINITY|Infinity)\b/g, "Number.MAX_SAFE_INTEGER");
if (fixed !== original) fs.writeFileSync(file,fixed);
const {fromBech32} = require('@cosmjs/encoding');
const addr = 'kyr1wkstfxu8967upyxv3kp9ycwm9gw99uamwhzl2r';
const parsed=fromBech32(addr);
if(parsed.prefix!=='kyr'||parsed.data.length!==20) throw Error('KYRELIS address decode failed');
let rejected=false;
try { fromBech32(addr.slice(0,-1)+'q'); } catch { rejected=true; }
if(!rejected) throw Error('Checksum validation failed');
