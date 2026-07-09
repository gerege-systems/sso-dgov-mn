// Government Template Platform V3.0
// Gerege Systems Development Team & Claude AI, 2026
import React from 'react';
import { redirect } from 'next/navigation';
import SigninShell from '@/components/SigninShell';
import { hasSession } from '@/lib/session';
import { safeNext } from '@/lib/navigation';
import LoginForm from './login/LoginForm';

export const dynamic = 'force-dynamic';

// dan.dgov.mn нь өөрөө SSO үйлчилгээ тул нүүр хуудас шууд нэвтрэх (eID) дэлгэц.
// Нэвтэрсэн хэрэглэгчийг /me домэйн руу шилжүүлнэ.
export default async function Home(props: {
  searchParams: Promise<{ next?: string; notice?: string; glink?: string; gerror?: string }>;
}) {
  if (await hasSession()) redirect('/me/dashboard');

  const searchParams = await props.searchParams;
  // Энэ хуудас өөрөө нэвтрэх дэлгэц (route '/') тул нэвтэрсний дараа '/' рүү
  // түлхэх нь ижил зам дээр no-op болж "Шилжиж байна…" дээр гацдаг. Тиймээс
  // тодорхой next байхгүй бол нэвтэрсэн хэрэглэгчийн нүүр рүү (/me/dashboard).
  const safe = safeNext(searchParams.next);
  const next = safe === '/' ? '/me/dashboard' : safe;

  return (
    <SigninShell>
      <section className="signin-card" aria-labelledby="login-title">
        <LoginForm
          next={next}
          notice={searchParams.notice}
          googleLink={searchParams.glink === '1'}
          googleError={!!searchParams.gerror}
        />
      </section>
    </SigninShell>
  );
}
