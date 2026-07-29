import React from 'react';
import SigninShell from '@gerege/ui-core/components/SigninShell';
import LoginPanel from '@gerege/ui-core/components/LoginPanel';
import { safeNext } from '@gerege/ui-core/lib/navigation';
import { fetchAuthMode } from '@gerege/ui-core/lib/authMode';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';

export const metadata = { title: pageTitle('Нэвтрэх') };

// Нэвтрэх гадаргуу нь кодоор БИШ, backend-ийн AUTH_MODE-оор тодорхойлогдоно.
// Энэ платформ нь өөрөө танилтын үйлчилгээ тул `provider` (нэвтрэх карт энд
// гарна) — гэхдээ тэр нь энэ файлын мэдэх зүйл биш: ижил код `client` горимд
// дээд SSO руу шилжүүлэх товч үзүүлнэ. Ингэснээр SSO үйлчилгээ ба түүний
// хэрэглэгч платформ хоёр НЭГ кодтой болов.
export default async function LoginPage(props: {
  searchParams: Promise<{
    next?: string;
    notice?: string;
    glink?: string;
    gerror?: string;
    mfa?: string;
    error?: string;
  }>;
}) {
  const searchParams = await props.searchParams;
  const next = safeNext(searchParams.next);
  const auth = await fetchAuthMode();

  return (
    <SigninShell>
      <section className="signin-card" aria-labelledby="login-title">
        <LoginPanel
          auth={auth}
          next={next}
          notice={searchParams.notice}
          googleLink={searchParams.glink === '1'}
          googleError={!!searchParams.gerror}
          mfaGate={searchParams.mfa === '1'}
          ssoFailed={searchParams.error === 'sso'}
        />
      </section>
    </SigninShell>
  );
}
