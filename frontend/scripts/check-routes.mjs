// BFF route-ын бүрхүүл бүрэн эсэхийг шалгана.
//
// Route-уудын ЛОГИК нь `@gerege/ui-core/api/**`-д, харин Next.js route-ыг
// файлын системээр бүртгэдэг тул апп нь зам бүрд нэг мөрийн бүрхүүл барина.
// Багц шинэ route нэмэхэд апп бүрхүүл үүсгэхгүй бол тэр endpoint ЧИМЭЭГҮЙ
// алга болно — кодыг нь апп дотор харахгүй тул анзаарагдахгүй.
//
// Иймд: багцын route бүрд бүрхүүл байх ЁСТОЙ. Зориудаар гаргахгүй бол доорх
// EXCLUDE-д бич — зөрүү нь ил, шийдвэртэй байна.

import { readdirSync, statSync, existsSync, readFileSync } from 'node:fs';
import { join, relative } from 'node:path';

// Энэ платформ дээр зориудаар нээхгүй route-ууд.
//
// Route-ын жагсаалт нь аюулгүй байдлын ЗӨВШӨӨРЛИЙН жагсаалт — хөтчөөс
// backend-ийн аль зам хүрэхийг тодорхойлно. Иймд UI-гүй модулийн BFF-ийг
// нээх нь proxy гадаргууг шалтгаангүй өргөтгөнө. Доорхи бүлгүүд нь template
// платформын чадвар бөгөөд энэ репод харгалзах хуудас БАЙХГҮЙ:
const EXCLUDE = [
  // Нүүрний нээлттэй AI чат (LandingChat) — энэ платформын landing-д чат
  // байхгүй тул нэвтрэлтгүй AI гадаргууг огт нээхгүй. Route-ын жагсаалт нь
  // аюулгүй байдлын ЗӨВШӨӨРЛИЙН жагсаалт тул UI-гүй прокси нээх нь
  // шалтгаангүй.
  'public/ai/chat',
  'public/ai/chat/stream',
  'public/ai/tts',
];

// Зориудаар АППАД хэрэгжүүлсэн route-ууд — бүрхүүл биш, бүтэн код. Платформ
// тусгайлсан утга агуулдаг тохиолдолд (ж: `aasa` доторх iOS bundle ID).
const LOCAL = [];

const PKG = 'node_modules/@gerege/ui-core/src/api';
const APP = 'src/app/api';

function walk(dir, base = dir) {
  if (!existsSync(dir)) return [];
  return readdirSync(dir).flatMap((e) => {
    const p = join(dir, e);
    return statSync(p).isDirectory() ? walk(p, base)
      : p.endsWith('.ts') ? [relative(base, p).replace(/\.ts$/, '')] : [];
  });
}

if (!existsSync(PKG)) {
  console.error('✗ @gerege/ui-core суугаагүй байна.');
  process.exit(1);
}

const pkgRoutes = walk(PKG);
const missing = pkgRoutes.filter(
  (r) => !EXCLUDE.includes(r) && !LOCAL.includes(r) && !existsSync(join(APP, r, 'route.ts')),
);

// Бүрхүүл нь дахин экспортоос өөр юм агуулж эхэлбэл логик апп руу буцаж
// хуулагдаж эхэлсэн гэсэн үг — багцын утга алдагдана.
const fat = walk(APP)
  .filter((r) => r.endsWith('/route'))
  .filter((r) => !LOCAL.includes(r.replace(/\/route$/, '')))
  .map((r) => join(APP, `${r}.ts`))
  .filter((p) => {
    // Хүлээгдэх хэлбэр: нэг дахин экспорт + route-ын тохиргооны литерал мөрүүд.
    // Тохиргоог ЛИТЕРАЛААР бичнэ — Turbopack нь дахин экспортолсон
    // `export const dynamic`-ыг статикаар уншиж чаддаггүй.
    const body = readFileSync(p, 'utf8').split('\n').filter((l) => l.trim());
    if (!body[0]?.startsWith('export {')) return true;
    return body.slice(1).some((l) => !/^export const \w+ = .+;$/.test(l.trim()));
  });

if (missing.length) {
  console.error(`✗ Багцын ${missing.length} route-д бүрхүүл алга:`);
  for (const r of missing) console.error(`    src/app/api/${r}/route.ts`);
  console.error('  Бүрхүүл үүсгэ, эсвэл scripts/check-routes.mjs-ийн EXCLUDE-д нэм.');
}
if (fat.length) {
  console.error(`✗ ${fat.length} бүрхүүл дахин экспортоос өөр юм агуулж байна:`);
  for (const p of fat) console.error(`    ${p}`);
}
if (missing.length || fat.length) process.exit(1);

const localNote = LOCAL.length ? `, ${LOCAL.length} нь аппад зориудаар` : '';
console.log(`✓ ${pkgRoutes.length} BFF route бүрхүүлтэй${localNote}.`);
