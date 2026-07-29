// Gerege Systems Development Team & Claude AI, 2026
import React from 'react';
import { redirect } from 'next/navigation';
import { hasSession } from '@gerege/ui-core/lib/session';
import { safeNext } from '@gerege/ui-core/lib/navigation';
import { fetchActiveTheme } from '@gerege/ui-core/lib/api';
import { fetchAuthMode, loginHref } from '@gerege/ui-core/lib/authMode';
import LoginPanel from '@gerege/ui-core/components/LoginPanel';
import LandingPage from '@/components/landing/LandingPage';

export const dynamic = 'force-dynamic';

// Нүүр хуудас нь платформын чадваруудыг харуулсан landing. Нэвтрэх гадаргуу нь
// backend-ийн AUTH_MODE-оос хамаарна:
//
//   provider — hero дотор нэвтрэх картыг шигтгэнэ, товчнууд #login руу;
//   client   — карт байхгүй, товчнууд дээд SSO руу шилжүүлнэ.
//
// Энэ платформ нь өөрөө танилтын үйлчилгээ тул production-д `provider`, гэхдээ
// код нь горимоос хамаарахгүй — яг ижил build хоёуланд ажиллана.
//
// Нэвтэрсэн хэрэглэгчийг /me домэйн руу шилжүүлнэ.
export default async function Home(props: {
  searchParams: Promise<{ next?: string; notice?: string; glink?: string; gerror?: string }>;
}) {
  if (await hasSession()) redirect('/me/dashboard');

  const searchParams = await props.searchParams;
  // Нэвтрэх карт энэ хуудсан дээр байж БОЛНО (route '/') тул нэвтэрсний дараа
  // '/' рүү түлхэх нь ижил зам дээр no-op болж гацдаг. Тиймээс тодорхой next
  // байхгүй бол нэвтэрсэн хэрэглэгчийн нүүр рүү (/me/dashboard).
  const safe = safeNext(searchParams.next);
  const next = safe === '/' ? '/me/dashboard' : safe;

  // Идэвхтэй theme-ийн landing текст/цэс — LandingPage copy.ts default дээр merge хийнэ.
  const [theme, auth] = await Promise.all([fetchActiveTheme(), fetchAuthMode()]);

  return (
    <LandingPage
      loginHref={loginHref(auth, next)}
      loginSlot={
        auth.mode === 'provider' ? (
          <LoginPanel
            auth={auth}
            next={next}
            notice={searchParams.notice}
            googleLink={searchParams.glink === '1'}
            googleError={!!searchParams.gerror}
            // Нүүр дээр автомат фокус хийхгүй — энэ нь маркетингийн хуудас
            // бөгөөд карт нь mobile-д hero-гийн доор байрладаг.
            autoFocus={false}
          />
        ) : null
      }
      themeLanding={theme.landing}
    />
  );
}
