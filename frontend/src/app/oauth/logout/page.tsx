// eID based AI enabled Government Template Platform V3.0
// OIDC provider RP-initiated logout — Hydra нь logout_challenge-тэй энд чиглүүлнэ.
import OAuthLogoutClient from './OAuthLogoutClient';

export const dynamic = 'force-dynamic';

export default async function OAuthLogoutPage(props: {
  searchParams: Promise<{ logout_challenge?: string }>;
}) {
  const { logout_challenge: challenge } = await props.searchParams;
  return <OAuthLogoutClient challenge={challenge ?? ''} />;
}
