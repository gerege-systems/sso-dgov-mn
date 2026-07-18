import React from 'react';
import SigninShell from '@/components/SigninShell';
import { safeNext } from '@/lib/navigation';
import LoginForm from './LoginForm';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Нэвтрэх — Төрийн нэгдсэн нэвтрэлт' };

export default async function LoginPage(
  props: {
    searchParams: Promise<{ next?: string; notice?: string; glink?: string; gerror?: string; mfa?: string }>;
  }
) {
  const searchParams = await props.searchParams;
  const next = safeNext(searchParams.next);

  return (
    <SigninShell>
      <section className="signin-card" aria-labelledby="login-title">
        <LoginForm
          next={next}
          notice={searchParams.notice}
          googleLink={searchParams.glink === '1'}
          googleError={!!searchParams.gerror}
          mfaGate={searchParams.mfa === '1'}
        />
      </section>
    </SigninShell>
  );
}
