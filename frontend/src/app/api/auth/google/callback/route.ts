import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';
import { backendFetch } from '@/lib/api';
import { setSession } from '@/lib/session';
import { cookieOptions } from '@/lib/cookies';

export const dynamic = 'force-dynamic';

// GET /api/auth/google/callback — Google OAuth redirect-ийн хүлээн авагч.
// state-ийг cookie-той тулгаж (CSRF), code-ийг backend руу дамжуулж token
// exchange хийлгэнэ. Backend:
//   - linked=true → токен хосыг cookie-д суулгаад /me/dashboard руу шилжүүлнэ.
//   - linked=false (эхний удаа) → link_token-ийг g_link cookie-д хадгалаад
//     /login?glink=1 руу шилжүүлж, eID-ээр баталгаажуулахыг хүсэнэ.
interface GoogleData {
  linked: boolean;
  user?: { token?: string; refresh_token?: string };
  link_token?: string;
  email?: string;
}

export async function GET(req: Request) {
  const url = new URL(req.url);
  const origin = process.env.APP_ORIGIN ?? url.origin;
  const code = url.searchParams.get('code');
  const state = url.searchParams.get('state');
  const savedState = (await cookies()).get('g_oauth_state')?.value;
  (await cookies()).delete('g_oauth_state');

  if (url.searchParams.get('error') || !code) {
    return NextResponse.redirect(`${origin}/login?gerror=google_cancelled`);
  }
  if (!state || !savedState || state !== savedState) {
    return NextResponse.redirect(`${origin}/login?gerror=state_mismatch`);
  }

  const redirectUri = `${origin}/api/auth/google/callback`;
  const result = await backendFetch<GoogleData>('/auth/google', {
    method: 'POST',
    body: JSON.stringify({ code, redirect_uri: redirectUri }),
  });

  if (!result.ok || !result.data) {
    return NextResponse.redirect(`${origin}/login?gerror=google_failed`);
  }

  const data = result.data;

  // Аль хэдийн холбогдсон → шууд нэвтрүүлнэ.
  if (data.linked && data.user?.token && data.user?.refresh_token) {
    await setSession(data.user.token, data.user.refresh_token);
    return NextResponse.redirect(`${origin}/me/dashboard`);
  }

  // Эхний удаа → link_token-ийг богино хугацааны cookie-д хадгалаад eID алхам руу.
  if (data.link_token) {
    (await cookies()).set('g_link', data.link_token, { ...cookieOptions(900), maxAge: 900 }); // 15 мин
    return NextResponse.redirect(`${origin}/login?glink=1`);
  }

  return NextResponse.redirect(`${origin}/login?gerror=google_failed`);
}
