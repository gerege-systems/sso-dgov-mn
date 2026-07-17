import React from 'react';
import { Inter, JetBrains_Mono, Source_Serif_4 } from 'next/font/google';
import './globals.css';
import { LangProvider } from '@/lib/lang';
import Providers from '@/components/Providers';

// Фонтыг build үед татаж next/font өөрөө host хийдэг тул CSP-г чанд 'self'-ээр
// үлдээж болно (гадны фонт host хэрэггүй).
const inter = Inter({
  subsets: ['latin'],
  weight: ['400', '500', '600', '700'],
  variable: '--font-display-stack',
  display: 'swap',
});

const interBody = Inter({
  subsets: ['latin'],
  weight: ['400', '500', '600'],
  variable: '--font-body-stack',
  display: 'swap',
});

const jbMono = JetBrains_Mono({
  subsets: ['latin'],
  weight: ['400', '500'],
  variable: '--font-mono-stack',
  display: 'swap',
});

// Сонголттой serif гэр бүл — html[data-font="serif"] үед --font-serif-stack-аар
// display/body-д тэжээгдэнэ. next/font build-time host хийдэг тул CSP 'self' хэвээр.
const sourceSerif = Source_Serif_4({
  subsets: ['latin'],
  weight: ['400', '500', '600'],
  variable: '--font-serif-stack',
  display: 'swap',
});

export const metadata = {
  title: 'DAN-Government SSO',
  description:
    'eID based, AI enabled. DAN-Government SSO — chi (net/http) + pgx дээр суурилсан төрийн нэгдсэн нэвтрэлт (SSO): eID-ээр нэвтрэх, профайл болон аюулгүй байдлын тохиргоог нэг дороос.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html
      lang="mn"
      className={`${inter.variable} ${interBody.variable} ${jbMono.variable} ${sourceSerif.variable}`}
      // theme-bootstrap.js нь hydration-аас өмнө <html>-д data-theme-pref
      // тавьдаг тул server/client attribute зөрүүгийн warning-ийг дарна.
      suppressHydrationWarning
    >
      <head>
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <meta name="color-scheme" content="light dark" />
        <link rel="icon" type="image/webp" href="/brand.webp" />
        {/* FOUC-аас сэргийлэх — гадаад блоклогч script (public/theme-bootstrap.js).
            Статик, адил-origin, 0.5KB файл тул XSS / гуравдагч талын эрсдэлгүй;
            body зурахаас ӨМНӨ ажиллах ёстой тул async/defer хийхгүй (эс бөгөөс
            загвар анивчина). Иймд no-sync-scripts дүрмийг энд зориуд унтраав. */}
        {/* eslint-disable-next-line @next/next/no-sync-scripts */}
        <script src="/theme-bootstrap.js" />
      </head>
      <body><Providers><LangProvider>{children}</LangProvider></Providers></body>
    </html>
  );
}
