import React from 'react';
import { redirect } from 'next/navigation';
import ProfileView from '@gerege/ui-core/components/me/ProfileView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Профайл') };

export default async function MeProfilePage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/me/profile');
  return <ProfileView me={me} />;
}
