import React from 'react';
import SigninShell from '@/components/SigninShell';
import { safeNext } from '@/lib/navigation';
import LoginForm from './LoginForm';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Нэвтрэх — Government Template v3.0' };

export default async function LoginPage(
  props: {
    searchParams: Promise<{ next?: string; notice?: string; glink?: string; gerror?: string }>;
  }
) {
  const searchParams = await props.searchParams;
  const next = safeNext(searchParams.next);

  return (
    <SigninShell>
      <section className="signin-card signin-card--narrow" aria-labelledby="login-title">
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
