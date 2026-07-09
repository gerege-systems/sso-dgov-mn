// Иргэний "Төрийн үйлчилгээ" порталын BFF (/api/gov/*) хариунуудын TS хэлбэрүүд.
// Backend-ийн responses_gov.go-той тэнцүү (json tag-аар).

export interface GovService {
  id: string;
  code: string;
  name: string;
  category: string;
  agency: string;
  description: string;
  fee: number;
  processing_days: number;
  online: boolean;
}

export type GovStatus =
  | 'submitted' | 'in_review' | 'approved' | 'rejected' | 'completed' | 'cancelled';

export interface GovApplication {
  id: string;
  service_name: string;
  reference_no: string;
  status: GovStatus;
  note: string;
  submitted_at: string;
  updated_at: string | null;
}

export interface GovReference {
  id: string;
  type: string;
  title: string;
  reference_no: string;
  status: string;
  issued_at: string;
  valid_until: string | null;
  data: unknown;
}

export interface GovNotification {
  id: string;
  title: string;
  body: string;
  category: 'info' | 'success' | 'warning';
  read: boolean;
  created_at: string;
}

export interface GovPayment {
  id: string;
  title: string;
  category: 'tax' | 'fee' | 'fine';
  amount: number;
  currency: string;
  status: 'pending' | 'paid';
  due_date: string | null;
  paid_at: string | null;
  created_at: string;
}

export interface GovAppointment {
  id: string;
  service_name: string;
  agency: string;
  location: string;
  scheduled_at: string;
  status: 'booked' | 'confirmed' | 'cancelled' | 'completed';
  note: string;
}

export interface GovOverview {
  open_applications: number;
  unread_notifications: number;
  unpaid_count: number;
  unpaid_amount: number;
  upcoming_count: number;
  issued_references: number;
  recent_applications: GovApplication[];
  upcoming_appointments: GovAppointment[];
}
