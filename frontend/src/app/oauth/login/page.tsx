// eID based AI enabled Government Template Platform V3.0
// OIDC provider login хуудас — Hydra нь browser-ыг энд login_challenge-тэй
// чиглүүлнэ. Иргэн нэвтэрсэн бол challenge-ыг шууд accept хийж (subject = dan
// user ID) Hydra руу буцна; эс бөгөөс dan-ийн /login руу (?next-ээр буцаж ирнэ).
import { redirect } from 'next/navigation';
import { getAccessToken } from '@/lib/session';
import OAuthLoginClient from './OAuthLoginClient';

export const dynamic = 'force-dynamic';

export default async function OAuthLoginPage(props: {
  searchParams: Promise<{ login_challenge?: string }>;
}) {
  const { login_challenge: challenge } = await props.searchParams;
  if (!challenge) redirect('/');
  const token = await getAccessToken();
  if (!token) {
    const ret = `/oauth/login?login_challenge=${encodeURIComponent(challenge)}`;
    redirect(`/login?next=${encodeURIComponent(ret)}`);
  }
  return <OAuthLoginClient challenge={challenge} />;
}
