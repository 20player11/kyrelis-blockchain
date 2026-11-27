const menu = document.querySelector('.hub-menu');
if (location.hostname === 'blockchain.tail3fb8f3.ts.net') {
  menu.hidden = false;
  for (const link of menu.querySelectorAll('a')) {
    if (link.pathname === '/' && link.origin === location.origin) link.setAttribute('aria-current', 'page');
  }
  document.addEventListener('click', event => {
    if (!menu.contains(event.target)) menu.open = false;
  });
  menu.addEventListener('keydown', event => {
    if (event.key === 'Escape') {
      menu.open = false;
      menu.querySelector('summary').focus();
    }
  });
}
if (document.querySelector('.explorer')) {
  let theme = 'auto';
  try {theme = JSON.parse(localStorage.getItem(import.meta.env.VITE_CHAIN_ID + '/desk/theme')) || theme;} catch {}
  const light = matchMedia('(prefers-color-scheme: light)');
  function apply() { document.documentElement.dataset.theme = (theme === 'light' || (theme === 'auto' && light.matches)) ? 'light' : 'dark'; }
  light.addEventListener('change', apply);
  apply();
}
