import React from 'react';
import SigninShell from '@gerege/ui-core/components/SigninShell';
import { safeNext } from '@gerege/ui-core/lib/navigation';
import LoginForm from '@gerege/ui-core/components/LoginForm';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';

export const metadata = { title: pageTitle('Супер админ нэвтрэх') };

// Superadmin нэвтрэлт — Google / eID сонголт нь ердийн LoginForm-той адил.
// MFA-той superadmin бол backend session-ий оронд MFA gate буцаана; LoginForm
// (eID poll) эсвэл Google callback (?mfa=1) TOTP/recovery challenge руу шилжүүлнэ.
// Амжилттай нэвтэрмэгц /admin/dashboard руу.
export default async function SuperadminLoginPage(props: {
  searchParams: Promise<{ next?: string; mfa?: string; gerror?: string }>;
}) {
  const searchParams = await props.searchParams;
  const next = safeNext(searchParams.next ?? '/admin/dashboard');

  return (
    <SigninShell>
      <section className="signin-card" aria-labelledby="login-title">
        <LoginForm
          next={next}
          googleError={!!searchParams.gerror}
          mfaGate={searchParams.mfa === '1'}
        />
      </section>
    </SigninShell>
  );
}
