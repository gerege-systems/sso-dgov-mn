// eID based AI enabled Government Template Platform V3.0
// OIDC provider (RP-facing) login хуудас — Hydra нь browser-ыг энд login_challenge-
// тэй чиглүүлнэ. dan-ий ӨӨРИЙН дизайнаар (SigninShell + LoginForm: eID РД/QR +
// Google) нэвтрүүлж, буцаж ирэхэд challenge-ыг accept хийнэ. Нэвтэрсэн сесстэй бол
// шууд accept.
import { redirect } from 'next/navigation';
import { getAccessToken } from '@/lib/session';
import LoginForm from '@/app/login/LoginForm';
import AcceptClient from './AcceptClient';

export const dynamic = 'force-dynamic';

export default async function OAuthLoginPage(props: {
  searchParams: Promise<{ login_challenge?: string; glink?: string; gerror?: string }>;
}) {
  const sp = await props.searchParams;
  const challenge = sp.login_challenge;
  if (!challenge) redirect('/');
  const hasSession = !!(await getAccessToken());
  const next = `/oauth/login?login_challenge=${challenge}`;

  return (
    <section className="signin-card signin-card--narrow" aria-labelledby="login-title">
      {hasSession ? (
        <AcceptClient challenge={challenge} />
      ) : (
        <LoginForm next={next} googleLink={sp.glink === '1'} googleError={!!sp.gerror} />
      )}
    </section>
  );
}
