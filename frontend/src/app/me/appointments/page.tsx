import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import GovAppointmentsView from '@gerege/ui-core/components/gov/GovAppointmentsView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Цаг захиалга') };

export default async function MeAppointmentsPage() {
  const me = await fetchMe();
  if (!me) redirect('/');
  return (
    <>
      <PageHead eyebrowKey="group.govServices" titleKey="nav.govAppointments" subKey="gov.appointments.sub" />
      <GovAppointmentsView />
    </>
  );
}
