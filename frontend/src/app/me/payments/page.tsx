import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import GovPaymentsView from '@gerege/ui-core/components/gov/GovPaymentsView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Төлбөр') };

export default async function MePaymentsPage() {
  const me = await fetchMe();
  if (!me) redirect('/');
  return (
    <>
      <PageHead eyebrowKey="group.govServices" titleKey="nav.govPayments" subKey="gov.payments.sub" />
      <GovPaymentsView />
    </>
  );
}
