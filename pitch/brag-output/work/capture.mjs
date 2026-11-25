// uso: node capture.mjs stills 1.2 3 ...   |   node capture.mjs all
import { chromium } from 'playwright'; // npm i playwright; usa o Google Chrome instalado (channel 'chrome')
import { pathToFileURL } from 'url';
import fs from 'fs';
const dir = new URL('.', import.meta.url).pathname;
const [mode, ...ts] = process.argv.slice(2);
const b = await chromium.launch({ channel: 'chrome' });
const p = await b.newPage({ viewport: { width: 1920, height: 1080 } });
p.on('pageerror', e => console.log('ERR', e.message));
await p.goto(pathToFileURL(dir + 'render.html').href);
await p.evaluate(() => document.fonts.ready);
const out = dir + (mode === 'all' ? 'out' : 'stills'); fs.mkdirSync(out, { recursive: true });
const times = mode === 'all' ? [...Array(1200).keys()].map(i => i / 30) : ts.map(Number);
for (const [i, t] of times.entries()) {
  await p.evaluate(t => window.render(t), t);
  const name = mode === 'all' ? `${String(i).padStart(4, '0')}.png` : `t${t.toFixed(2)}.png`;
  await p.screenshot({ path: `${out}/${name}`, type: 'png' });
  if (mode === 'all' && i % 150 === 0) console.log(i);
}
await b.close();
