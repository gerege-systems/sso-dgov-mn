// eID based AI enabled Government Template Platform V3.0
// OIDC provider (RP-facing) login хуудас — Hydra нь browser-ыг энд login_challenge-
// тэй чиглүүлнэ. dan-ий өөрийн апп login (/login)-оос ТУСДАА, sso.dgov.mn-ий UI-
// гаар нэгдсэн нэвтрэлтийн дэлгэц харуулж, eID-ээр баталгаажуулаад challenge-ыг
// accept хийнэ. Нэвтэрсэн сесстэй бол дахин eID шаардахгүй шууд accept.
import { redirect } from 'next/navigation';
import { getAccessToken } from '@/lib/session';
import OAuthLoginClient from './OAuthLoginClient';

export const dynamic = 'force-dynamic';

export default async function OAuthLoginPage(props: {
  searchParams: Promise<{ login_challenge?: string }>;
}) {
  const { login_challenge: challenge } = await props.searchParams;
  if (!challenge) redirect('/');
  const hasSession = !!(await getAccessToken());
  return <OAuthLoginClient challenge={challenge} hasSession={hasSession} />;
}
