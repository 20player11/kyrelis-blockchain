import {defineConfig} from 'vite';
const chain = path => ({ws:path==='/rpc',target:'http://127.0.0.1:'+ (path === '/rpc' ? 26657 : 1317), rewrite:url=>url.slice(path.length)||'/'});
export default defineConfig({build:{rollupOptions:{input:{hub:'index.html',explorer:'explorer.html'}}},preview:{host:'127.0.0.1',port:5173,strictPort:true,proxy:{'/api':'http://127.0.0.1:18188','/rpc':chain('/rpc'),'/rest':chain('/rest')}}});
