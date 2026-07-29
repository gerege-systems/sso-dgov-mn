import React from 'react';
import { redirect } from 'next/navigation';
import HomeView from '@gerege/ui-core/components/me/HomeView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Хяналтын самбар') };

export default async function MeDashboardPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/me/dashboard');
  return <HomeView me={me} />;
}
