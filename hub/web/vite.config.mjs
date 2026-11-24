import {defineConfig} from 'vite';
const chain = path => ({target:'http://127.0.0.1:'+ (path === '/rpc' ? 26657 : 1317), rewrite:url=>url.slice(path.length)||'/'});
export default defineConfig({preview:{host:'127.0.0.1',port:5173,strictPort:true,proxy:{'/api':'http://127.0.0.1:18188','/rpc':chain('/rpc'),'/rest':chain('/rest')}}});
