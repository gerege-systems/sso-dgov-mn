-- Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.
--
-- НЭГДСЭН СУУРЬ SCHEMA (consolidated init). Хуучин 1..37 migration-уудыг DB-ийн
-- эцсийн төлвөөс pg_dump-аар нэгтгэв (шинэ суулгалтад ганц энэ файл гүйнэ; хуучин
-- migration-ууд migrations/old/-д архивлагдсан). Seed data (roles/permissions/
-- gateway/ai/themes/gov_services) + RLS policy бүгд багтсан. app_user-ийн privilege
-- нарийсгалт (хуучин migration 17) төгсгөлд (app_user байхгүй бол no-op).
--

--
-- PostgreSQL database dump
--

-- Dumped from database version 16.14
-- Dumped by pg_dump version 16.14

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', 'public', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;

--
-- Name: app_is_org_member(uuid, uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.app_is_org_member(p_org_id uuid, p_user_id uuid) RETURNS boolean
    LANGUAGE sql STABLE SECURITY DEFINER
    AS $$
    SELECT p_user_id IS NOT NULL AND EXISTS (
        SELECT 1 FROM organization_memberships
        WHERE org_id = p_org_id AND user_id = p_user_id
    );
$$;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admin_api_keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admin_api_keys (
    id text NOT NULL,
    name text NOT NULL,
    hash text NOT NULL,
    display text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone,
    disabled boolean DEFAULT false NOT NULL
);

--
-- Name: ai_knowledge; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_knowledge (
    id integer NOT NULL,
    title text NOT NULL,
    content text NOT NULL,
    tags text[] DEFAULT '{}'::text[] NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone
);

--
-- Name: ai_knowledge_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_knowledge_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: ai_knowledge_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_knowledge_id_seq OWNED BY public.ai_knowledge.id;

--
-- Name: ai_prompts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_prompts (
    key text NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    updated_at timestamp with time zone
);

--
-- Name: application_services; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.application_services (
    application_id uuid NOT NULL,
    service_id uuid NOT NULL
);

--
-- Name: applications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.applications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    client_id text NOT NULL,
    name text NOT NULL,
    app_type text DEFAULT 'm2m'::text NOT NULL,
    tags text[] DEFAULT '{}'::text[] NOT NULL,
    redirect_uris text[] DEFAULT '{}'::text[] NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    created_by text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone
);

--
-- Name: audit_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_log (
    id bigint NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    actor_user_id uuid,
    action text NOT NULL,
    category text,
    target text,
    request_id text,
    metadata jsonb,
    prev_hash text,
    chain_hash text NOT NULL
);

ALTER TABLE ONLY public.audit_log FORCE ROW LEVEL SECURITY;

--
-- Name: audit_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.audit_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: audit_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.audit_log_id_seq OWNED BY public.audit_log.id;

--
-- Name: developer_apps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.developer_apps (
    client_id text NOT NULL,
    owner_eid_sub text NOT NULL,
    name text NOT NULL,
    redirect_uris text[] DEFAULT '{}'::text[] NOT NULL,
    scopes text[] DEFAULT '{}'::text[] NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: gateway_request_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gateway_request_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    method text NOT NULL,
    path text NOT NULL,
    status integer NOT NULL,
    latency_ms integer DEFAULT 0 NOT NULL,
    client_ip text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: gateway_services; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gateway_services (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    protocol text DEFAULT 'https'::text NOT NULL,
    host text NOT NULL,
    port integer DEFAULT 443 NOT NULL,
    path text DEFAULT '/'::text NOT NULL,
    retries integer DEFAULT 3 NOT NULL,
    connect_timeout_ms integer DEFAULT 60000 NOT NULL,
    tags text[] DEFAULT '{}'::text[] NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone,
    scope text DEFAULT ''::text NOT NULL
);

--
-- Name: gov_applications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gov_applications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    service_id uuid,
    service_name text DEFAULT ''::text NOT NULL,
    reference_no text NOT NULL,
    status text DEFAULT 'submitted'::text NOT NULL,
    note text DEFAULT ''::text NOT NULL,
    submitted_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone
);

ALTER TABLE ONLY public.gov_applications FORCE ROW LEVEL SECURITY;

--
-- Name: gov_appointments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gov_appointments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    service_id uuid,
    service_name text DEFAULT ''::text NOT NULL,
    agency text DEFAULT ''::text NOT NULL,
    location text DEFAULT ''::text NOT NULL,
    scheduled_at timestamp with time zone NOT NULL,
    status text DEFAULT 'booked'::text NOT NULL,
    note text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.gov_appointments FORCE ROW LEVEL SECURITY;

--
-- Name: gov_notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gov_notifications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    title text NOT NULL,
    body text DEFAULT ''::text NOT NULL,
    category text DEFAULT 'info'::text NOT NULL,
    read boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.gov_notifications FORCE ROW LEVEL SECURITY;

--
-- Name: gov_payments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gov_payments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    title text NOT NULL,
    category text DEFAULT 'fee'::text NOT NULL,
    amount integer DEFAULT 0 NOT NULL,
    currency text DEFAULT 'MNT'::text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    due_date timestamp with time zone,
    paid_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.gov_payments FORCE ROW LEVEL SECURITY;

--
-- Name: gov_references; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gov_references (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    type text NOT NULL,
    title text NOT NULL,
    reference_no text NOT NULL,
    status text DEFAULT 'issued'::text NOT NULL,
    issued_at timestamp with time zone DEFAULT now() NOT NULL,
    valid_until timestamp with time zone,
    data jsonb DEFAULT '{}'::jsonb NOT NULL
);

ALTER TABLE ONLY public.gov_references FORCE ROW LEVEL SECURITY;

--
-- Name: gov_services; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gov_services (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    category text DEFAULT ''::text NOT NULL,
    agency text DEFAULT ''::text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    fee integer DEFAULT 0 NOT NULL,
    processing_days integer DEFAULT 0 NOT NULL,
    online boolean DEFAULT true NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: login_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.login_events (
    id bigint NOT NULL,
    eid_sub text,
    method text NOT NULL,
    national_id text,
    google_sub text,
    raw_claims jsonb DEFAULT '{}'::jsonb NOT NULL,
    ip text,
    user_agent text,
    is_new_device boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: login_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.login_events ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.login_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

--
-- Name: org_stamps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.org_stamps (
    org_register character varying(16) NOT NULL,
    image text NOT NULL,
    uploaded_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: organization_memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organization_memberships (
    org_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role text DEFAULT 'member'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.organization_memberships FORCE ROW LEVEL SECURITY;

--
-- Name: organizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organizations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    reg_no text NOT NULL,
    name text NOT NULL,
    name_latin text DEFAULT ''::text NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);

ALTER TABLE ONLY public.organizations FORCE ROW LEVEL SECURITY;

--
-- Name: permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permissions (
    key text NOT NULL,
    label text NOT NULL,
    category text DEFAULT ''::text NOT NULL
);

--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role_permissions (
    role_id integer NOT NULL,
    permission_key text NOT NULL
);

--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.roles (
    id integer NOT NULL,
    key text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    is_system boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone
);

--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.roles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;

--
-- Name: security_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.security_events (
    id bigint NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    user_id uuid,
    kind text NOT NULL,
    severity text,
    source text,
    user_agent text,
    ip text,
    detail jsonb
);

ALTER TABLE ONLY public.security_events FORCE ROW LEVEL SECURITY;

--
-- Name: security_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.security_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: security_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.security_events_id_seq OWNED BY public.security_events.id;

--
-- Name: site_appearance; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_appearance (
    id smallint DEFAULT 1 NOT NULL,
    accent text DEFAULT 'cobalt'::text NOT NULL,
    font text DEFAULT 'inter'::text NOT NULL,
    style text DEFAULT 'comfortable'::text NOT NULL,
    theme text DEFAULT 'light'::text NOT NULL,
    updated_at timestamp with time zone,
    CONSTRAINT site_appearance_id_check CHECK ((id = 1))
);

--
-- Name: superadmin_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.superadmin_accounts (
    user_id uuid NOT NULL,
    civil_id text,
    national_id text,
    email_verified boolean DEFAULT false NOT NULL,
    mfa_enabled boolean DEFAULT false NOT NULL,
    totp_secret text,
    invited_by text DEFAULT ''::text NOT NULL,
    onboarded_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone
);

ALTER TABLE ONLY public.superadmin_accounts FORCE ROW LEVEL SECURITY;

--
-- Name: superadmin_invites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.superadmin_invites (
    email text NOT NULL,
    invited_by text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    accepted_at timestamp with time zone
);

--
-- Name: themes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.themes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    config jsonb DEFAULT '{}'::jsonb NOT NULL,
    is_active boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone
);

--
-- Name: user_integrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_integrations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    provider text NOT NULL,
    access_token text NOT NULL,
    refresh_token text,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone
);

ALTER TABLE ONLY public.user_integrations FORCE ROW LEVEL SECURITY;

--
-- Name: user_recovery_codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_recovery_codes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    code_hash text NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.user_recovery_codes FORCE ROW LEVEL SECURITY;

--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    username character varying(25) NOT NULL,
    email character varying(50),
    password character varying(255),
    active boolean NOT NULL,
    role_id smallint NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    password_changed_at timestamp with time zone,
    first_name text DEFAULT ''::text NOT NULL,
    last_name text DEFAULT ''::text NOT NULL,
    first_name_en text DEFAULT ''::text NOT NULL,
    last_name_en text DEFAULT ''::text NOT NULL,
    national_id text,
    civil_id text,
    kyc_level text,
    document_number text,
    cert_serial text,
    cert_not_before timestamp with time zone,
    cert_not_after timestamp with time zone,
    cert_issuer text,
    cert_key_type text,
    google_sub text,
    google_email text,
    google_email_verified boolean DEFAULT false NOT NULL,
    google_name text,
    google_picture text,
    google_linked_at timestamp with time zone,
    signature_image text
);

ALTER TABLE ONLY public.users FORCE ROW LEVEL SECURITY;

--
-- Name: ai_knowledge id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_knowledge ALTER COLUMN id SET DEFAULT nextval('public.ai_knowledge_id_seq'::regclass);

--
-- Name: audit_log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_log ALTER COLUMN id SET DEFAULT nextval('public.audit_log_id_seq'::regclass);

--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);

--
-- Name: security_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.security_events ALTER COLUMN id SET DEFAULT nextval('public.security_events_id_seq'::regclass);

--
-- Data for Name: admin_api_keys; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: ai_knowledge; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.ai_knowledge VALUES (1, 'Нууц үг сэргээх', 'Нэвтрэх хуудасны «Нууц үгээ мартсан» холбоосоор имэйлээ оруулахад нэг удаагийн код очно. Кодоо оруулаад 12-оос доошгүй тэмдэгттэй, том/жижиг үсэг, тоо, тусгай тэмдэгт агуулсан шинэ нууц үг тохируулна.', '{"нууц үг",сэргээх,password}', '2026-07-18 10:30:30.862085+00', NULL);
INSERT INTO public.ai_knowledge VALUES (2, 'Бүртгэл идэвхжүүлэх', 'Бүртгүүлсний дараа имэйлээр очсон нэг удаагийн (OTP) кодыг баталгаажуулах хуудсанд оруулснаар бүртгэл идэвхжинэ. Код очоогүй бол дахин илгээх товчийг ашиглана.', '{бүртгэл,otp,идэвхжүүлэх}', '2026-07-18 10:30:30.862085+00', NULL);
INSERT INTO public.ai_knowledge VALUES (3, 'Эрхийн систем (RBAC)', 'Хэрэглэгч бүр нэг эрхтэй (админ, менежер, хэрэглэгч г.м.). Админ бүх эрхийг автоматаар эзэмшинэ; бусад эрхийн зөвшөөрлүүдийг админ Эрх (RBAC) хэсгээс тохируулна.', '{эрх,rbac,role}', '2026-07-18 10:30:30.862085+00', NULL);

--
-- Data for Name: ai_prompts; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.ai_prompts VALUES ('scope', 'Чи Gerege платформын албан ёсны туслах. Зөвхөн Gerege платформын үйлчилгээ, бүртгэл, нэвтрэлт, аюулгүй байдал, тохиргоо болон мэдлэгийн санд байгаа сэдвээр тусална.', NULL);
INSERT INTO public.ai_prompts VALUES ('instructions', '', NULL);

--
-- Data for Name: application_services; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.application_services VALUES ('80939679-5cad-43df-9892-b996e43d4d44', '185048e5-8edc-42d0-a153-fcb087f50bcb');
INSERT INTO public.application_services VALUES ('9a7e0c0d-1570-47d2-abe5-4172600c0186', '185048e5-8edc-42d0-a153-fcb087f50bcb');

--
-- Data for Name: applications; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.applications VALUES ('80939679-5cad-43df-9892-b996e43d4d44', 'template-dgov-mn', 'template.dgov.mn', 'web', '{rp}', '{https://template.dgov.mn/auth/callback}', true, 'seed-rp', '2026-07-18 10:30:31.238207+00', NULL);
INSERT INTO public.applications VALUES ('9a7e0c0d-1570-47d2-abe5-4172600c0186', 'developer-dgov-mn', 'developer.dgov.mn', 'web', '{rp,developer}', '{https://developer.dgov.mn/auth/callback}', true, 'seed-rp', '2026-07-18 10:30:31.238207+00', NULL);

--
-- Data for Name: audit_log; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: developer_apps; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: gateway_request_logs; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: gateway_services; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.gateway_services VALUES ('b8907e30-7aa5-4ff3-ab60-0029dc952e00', 'dan-sso', 'https', 'sso.dgov.mn', 443, '/oauth2', 3, 60000, '{sso,oidc}', true, '2026-07-18 10:30:31.237719+00', NULL, 'svc:dan-sso');
INSERT INTO public.gateway_services VALUES ('185048e5-8edc-42d0-a153-fcb087f50bcb', 'eid-sign', 'https', 'sso.dgov.mn', 443, '/rp/sign', 3, 60000, '{eid,sign}', true, '2026-07-18 10:30:31.237719+00', NULL, 'svc:eid-sign');

--
-- Data for Name: gov_applications; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: gov_appointments; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: gov_notifications; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: gov_payments; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: gov_references; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: gov_services; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.gov_services VALUES ('08d38891-0ffb-4fe9-a26d-0bdcb83b6457', 'CIVIL_ID', 'Иргэний үнэмлэх захиалах', 'Бүртгэл', 'УБЕГ', 'Иргэний үнэмлэх шинээр авах, дахин захиалах.', 25000, 7, true, true, '2026-07-18 10:30:31.172299+00');
INSERT INTO public.gov_services VALUES ('a349ad43-cecb-4ecd-ab1b-71f25b72f952', 'RESIDENCE', 'Оршин суугаа газрын лавлагаа', 'Бүртгэл', 'УБЕГ', 'Оршин суугаа хаягийн албан ёсны лавлагаа.', 500, 0, true, true, '2026-07-18 10:30:31.172299+00');
INSERT INTO public.gov_services VALUES ('7795c7b5-7c13-49f3-91e7-3617a1d7fa3e', 'TAX_CLEAR', 'Татварын тодорхойлолт', 'Татвар', 'ТЕГ', 'Татварын өргүй гэсэн тодорхойлолт.', 0, 1, true, true, '2026-07-18 10:30:31.172299+00');
INSERT INTO public.gov_services VALUES ('aaed9c8f-188f-4ebc-84cf-e83eb16c82ef', 'SOCIAL_INS', 'Нийгмийн даатгалын лавлагаа', 'Нийгмийн хамгаалал', 'НДЕГ', 'Шимтгэл төлөлтийн дэлгэрэнгүй лавлагаа.', 0, 0, true, true, '2026-07-18 10:30:31.172299+00');
INSERT INTO public.gov_services VALUES ('6be448d5-781e-46dd-b212-a6911b823b85', 'DRIVER_LIC', 'Жолооны үнэмлэх сунгах', 'Тээвэр', 'Зам тээвэр', 'Жолоочийн үнэмлэхний хугацаа сунгах.', 35000, 5, false, true, '2026-07-18 10:30:31.172299+00');
INSERT INTO public.gov_services VALUES ('7f73d254-9c64-4fa5-92c4-420ce1bc78e0', 'MARRIAGE', 'Гэрлэлтийн гэрчилгээ', 'Бүртгэл', 'УБЕГ', 'Гэрлэлт бүртгүүлэх, гэрчилгээ авах.', 15000, 3, false, true, '2026-07-18 10:30:31.172299+00');
INSERT INTO public.gov_services VALUES ('bf160bc5-6bc0-4af7-aad5-007a39fa5864', 'HEALTH_INS', 'Эрүүл мэндийн даатгал', 'Эрүүл мэнд', 'ЭМД', 'Эрүүл мэндийн даатгалын төлөв, төлбөр.', 0, 0, true, true, '2026-07-18 10:30:31.172299+00');
INSERT INTO public.gov_services VALUES ('d50ca682-bf39-4331-b92f-452d19eb1633', 'BIZ_REG', 'Аж ахуй нэгж бүртгэх', 'Бизнес', 'УБЕГ', 'ХХК/ХК шинээр бүртгүүлэх.', 44000, 10, true, true, '2026-07-18 10:30:31.172299+00');

--
-- Data for Name: login_events; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: org_stamps; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: organization_memberships; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: organizations; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: permissions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.permissions VALUES ('dashboard.view', 'Хяналтын самбар үзэх', 'general');
INSERT INTO public.permissions VALUES ('settings.manage', 'Тохиргоо удирдах', 'general');
INSERT INTO public.permissions VALUES ('users.manage', 'Хэрэглэгч удирдах', 'administration');
INSERT INTO public.permissions VALUES ('roles.manage', 'Эрх (role) удирдах', 'administration');
INSERT INTO public.permissions VALUES ('manager.view', 'Менежерийн хэсэг', 'management');
INSERT INTO public.permissions VALUES ('personal.view', 'Хувийн хэсэг', 'personal');
INSERT INTO public.permissions VALUES ('gateway.manage', 'API Gateway удирдах', 'administration');

--
-- Data for Name: role_permissions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.role_permissions VALUES (4, 'dashboard.view');
INSERT INTO public.role_permissions VALUES (4, 'personal.view');
INSERT INTO public.role_permissions VALUES (3, 'dashboard.view');
INSERT INTO public.role_permissions VALUES (3, 'manager.view');
INSERT INTO public.role_permissions VALUES (3, 'users.manage');

--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.roles VALUES (2, 'admin', 'Админ', 'Бүх эрхтэй системийн админ', true, '2026-07-18 10:30:30.775145+00', NULL);
INSERT INTO public.roles VALUES (4, 'user', 'Хэрэглэгч', 'Энгийн хэрэглэгч', true, '2026-07-18 10:30:30.775145+00', NULL);
INSERT INTO public.roles VALUES (3, 'manager', 'Менежер', 'Хэрэглэгч хянадаг менежер', true, '2026-07-18 10:30:30.775145+00', NULL);
INSERT INTO public.roles VALUES (1, 'superadmin', 'Супер админ', 'Админуудыг удирдах дээд эрх', true, '2026-07-18 10:30:31.267455+00', NULL);

--
-- Data for Name: security_events; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: site_appearance; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.site_appearance VALUES (1, 'cobalt', 'inter', 'comfortable', 'light', NULL);

--
-- Data for Name: superadmin_accounts; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: superadmin_invites; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: themes; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.themes VALUES ('fd6d1ec3-5519-4b91-b785-8a5c87a98ca4', 'DAN default', '{"landing": {"en": {}, "mn": {}}, "appearance": {"font": "inter", "mode": "light", "style": "comfortable", "colors": {}}}', true, '2026-07-18 10:30:31.415038+00', NULL);

--
-- Data for Name: user_integrations; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: user_recovery_codes; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Name: ai_knowledge_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.ai_knowledge_id_seq', 3, true);

--
-- Name: audit_log_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.audit_log_id_seq', 1, false);

--
-- Name: login_events_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.login_events_id_seq', 1, false);

--
-- Name: roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.roles_id_seq', 4, true);

--
-- Name: security_events_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.security_events_id_seq', 1, false);

--
-- Name: admin_api_keys admin_api_keys_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_api_keys
    ADD CONSTRAINT admin_api_keys_hash_key UNIQUE (hash);

--
-- Name: admin_api_keys admin_api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_api_keys
    ADD CONSTRAINT admin_api_keys_pkey PRIMARY KEY (id);

--
-- Name: ai_knowledge ai_knowledge_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_knowledge
    ADD CONSTRAINT ai_knowledge_pkey PRIMARY KEY (id);

--
-- Name: ai_prompts ai_prompts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_prompts
    ADD CONSTRAINT ai_prompts_pkey PRIMARY KEY (key);

--
-- Name: application_services application_services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.application_services
    ADD CONSTRAINT application_services_pkey PRIMARY KEY (application_id, service_id);

--
-- Name: applications applications_client_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applications
    ADD CONSTRAINT applications_client_id_key UNIQUE (client_id);

--
-- Name: applications applications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applications
    ADD CONSTRAINT applications_pkey PRIMARY KEY (id);

--
-- Name: audit_log audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_log
    ADD CONSTRAINT audit_log_pkey PRIMARY KEY (id);

--
-- Name: developer_apps developer_apps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.developer_apps
    ADD CONSTRAINT developer_apps_pkey PRIMARY KEY (client_id);

--
-- Name: gateway_request_logs gateway_request_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gateway_request_logs
    ADD CONSTRAINT gateway_request_logs_pkey PRIMARY KEY (id);

--
-- Name: gateway_services gateway_services_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gateway_services
    ADD CONSTRAINT gateway_services_name_key UNIQUE (name);

--
-- Name: gateway_services gateway_services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gateway_services
    ADD CONSTRAINT gateway_services_pkey PRIMARY KEY (id);

--
-- Name: gov_applications gov_applications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gov_applications
    ADD CONSTRAINT gov_applications_pkey PRIMARY KEY (id);

--
-- Name: gov_appointments gov_appointments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gov_appointments
    ADD CONSTRAINT gov_appointments_pkey PRIMARY KEY (id);

--
-- Name: gov_notifications gov_notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gov_notifications
    ADD CONSTRAINT gov_notifications_pkey PRIMARY KEY (id);

--
-- Name: gov_payments gov_payments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gov_payments
    ADD CONSTRAINT gov_payments_pkey PRIMARY KEY (id);

--
-- Name: gov_references gov_references_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gov_references
    ADD CONSTRAINT gov_references_pkey PRIMARY KEY (id);

--
-- Name: gov_services gov_services_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gov_services
    ADD CONSTRAINT gov_services_code_key UNIQUE (code);

--
-- Name: gov_services gov_services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gov_services
    ADD CONSTRAINT gov_services_pkey PRIMARY KEY (id);

--
-- Name: login_events login_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.login_events
    ADD CONSTRAINT login_events_pkey PRIMARY KEY (id);

--
-- Name: org_stamps org_stamps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.org_stamps
    ADD CONSTRAINT org_stamps_pkey PRIMARY KEY (org_register);

--
-- Name: organization_memberships organization_memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_memberships
    ADD CONSTRAINT organization_memberships_pkey PRIMARY KEY (org_id, user_id);

--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);

--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (key);

--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_key);

--
-- Name: roles roles_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_key_key UNIQUE (key);

--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);

--
-- Name: security_events security_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.security_events
    ADD CONSTRAINT security_events_pkey PRIMARY KEY (id);

--
-- Name: site_appearance site_appearance_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_appearance
    ADD CONSTRAINT site_appearance_pkey PRIMARY KEY (id);

--
-- Name: superadmin_accounts superadmin_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.superadmin_accounts
    ADD CONSTRAINT superadmin_accounts_pkey PRIMARY KEY (user_id);

--
-- Name: superadmin_invites superadmin_invites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.superadmin_invites
    ADD CONSTRAINT superadmin_invites_pkey PRIMARY KEY (email);

--
-- Name: themes themes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.themes
    ADD CONSTRAINT themes_pkey PRIMARY KEY (id);

--
-- Name: user_integrations user_integrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_integrations
    ADD CONSTRAINT user_integrations_pkey PRIMARY KEY (id);

--
-- Name: user_recovery_codes user_recovery_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_recovery_codes
    ADD CONSTRAINT user_recovery_codes_pkey PRIMARY KEY (id);

--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

--
-- Name: developer_apps_owner_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX developer_apps_owner_idx ON public.developer_apps USING btree (owner_eid_sub);

--
-- Name: idx_application_services_service; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_application_services_service ON public.application_services USING btree (service_id);

--
-- Name: idx_audit_log_actor; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_log_actor ON public.audit_log USING btree (actor_user_id);

--
-- Name: idx_audit_log_occurred_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_log_occurred_at ON public.audit_log USING btree (occurred_at);

--
-- Name: idx_gateway_request_logs_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_gateway_request_logs_created ON public.gateway_request_logs USING btree (created_at DESC);

--
-- Name: idx_gov_applications_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_gov_applications_user ON public.gov_applications USING btree (user_id, submitted_at DESC);

--
-- Name: idx_gov_appointments_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_gov_appointments_user ON public.gov_appointments USING btree (user_id, scheduled_at);

--
-- Name: idx_gov_notifications_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_gov_notifications_user ON public.gov_notifications USING btree (user_id, created_at DESC);

--
-- Name: idx_gov_payments_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_gov_payments_user ON public.gov_payments USING btree (user_id, created_at DESC);

--
-- Name: idx_gov_references_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_gov_references_user ON public.gov_references USING btree (user_id, issued_at DESC);

--
-- Name: idx_org_memberships_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_org_memberships_user ON public.organization_memberships USING btree (user_id);

--
-- Name: idx_organizations_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_organizations_created_by ON public.organizations USING btree (created_by);

--
-- Name: idx_organizations_reg_no_lower; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_organizations_reg_no_lower ON public.organizations USING btree (lower(reg_no)) WHERE (deleted_at IS NULL);

--
-- Name: idx_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_role_id ON public.users USING btree (role_id);

--
-- Name: idx_security_events_received_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_security_events_received_at ON public.security_events USING btree (received_at);

--
-- Name: idx_security_events_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_security_events_user ON public.security_events USING btree (user_id);

--
-- Name: idx_superadmin_accounts_civil; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_superadmin_accounts_civil ON public.superadmin_accounts USING btree (lower(civil_id)) WHERE (civil_id IS NOT NULL);

--
-- Name: idx_user_integrations_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_integrations_user ON public.user_integrations USING btree (user_id);

--
-- Name: idx_user_recovery_codes_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_recovery_codes_user ON public.user_recovery_codes USING btree (user_id);

--
-- Name: idx_users_civil_id_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_civil_id_active ON public.users USING btree (lower(civil_id)) WHERE (civil_id IS NOT NULL);

--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);

--
-- Name: idx_users_email_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email_active ON public.users USING btree (email) WHERE ((deleted_at IS NULL) AND (email IS NOT NULL) AND ((email)::text <> ''::text));

--
-- Name: idx_users_google_sub_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_google_sub_active ON public.users USING btree (google_sub) WHERE ((google_sub IS NOT NULL) AND (deleted_at IS NULL));

--
-- Name: idx_users_national_id_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_national_id_active ON public.users USING btree (lower(national_id)) WHERE (national_id IS NOT NULL);

--
-- Name: idx_users_username_active_ci; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_username_active_ci ON public.users USING btree (lower((username)::text)) WHERE (deleted_at IS NULL);

--
-- Name: login_events_eid_sub_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX login_events_eid_sub_idx ON public.login_events USING btree (eid_sub, created_at DESC);

--
-- Name: themes_one_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX themes_one_active ON public.themes USING btree (is_active) WHERE is_active;

--
-- Name: uq_user_integrations_user_provider; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_user_integrations_user_provider ON public.user_integrations USING btree (user_id, provider);

--
-- Name: application_services application_services_application_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.application_services
    ADD CONSTRAINT application_services_application_id_fkey FOREIGN KEY (application_id) REFERENCES public.applications(id) ON DELETE CASCADE;

--
-- Name: application_services application_services_service_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.application_services
    ADD CONSTRAINT application_services_service_id_fkey FOREIGN KEY (service_id) REFERENCES public.gateway_services(id) ON DELETE CASCADE;

--
-- Name: users fk_users_role; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES public.roles(id);

--
-- Name: gov_applications gov_applications_service_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gov_applications
    ADD CONSTRAINT gov_applications_service_id_fkey FOREIGN KEY (service_id) REFERENCES public.gov_services(id) ON DELETE SET NULL;

--
-- Name: gov_appointments gov_appointments_service_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gov_appointments
    ADD CONSTRAINT gov_appointments_service_id_fkey FOREIGN KEY (service_id) REFERENCES public.gov_services(id) ON DELETE SET NULL;

--
-- Name: organization_memberships organization_memberships_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_memberships
    ADD CONSTRAINT organization_memberships_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;

--
-- Name: organization_memberships organization_memberships_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_memberships
    ADD CONSTRAINT organization_memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: organizations organizations_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

--
-- Name: role_permissions role_permissions_permission_key_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_permission_key_fkey FOREIGN KEY (permission_key) REFERENCES public.permissions(key) ON DELETE CASCADE;

--
-- Name: role_permissions role_permissions_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;

--
-- Name: superadmin_accounts superadmin_accounts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.superadmin_accounts
    ADD CONSTRAINT superadmin_accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: user_integrations user_integrations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_integrations
    ADD CONSTRAINT user_integrations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: user_recovery_codes user_recovery_codes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_recovery_codes
    ADD CONSTRAINT user_recovery_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: audit_log; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.audit_log ENABLE ROW LEVEL SECURITY;

--
-- Name: audit_log audit_log_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY audit_log_admin ON public.audit_log USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: audit_log audit_log_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY audit_log_service ON public.audit_log USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: gov_applications; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.gov_applications ENABLE ROW LEVEL SECURITY;

--
-- Name: gov_applications gov_applications_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_applications_admin ON public.gov_applications USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: gov_applications gov_applications_self; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_applications_self ON public.gov_applications USING (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid))) WITH CHECK (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid)));

--
-- Name: gov_applications gov_applications_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_applications_service ON public.gov_applications USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: gov_appointments; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.gov_appointments ENABLE ROW LEVEL SECURITY;

--
-- Name: gov_appointments gov_appointments_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_appointments_admin ON public.gov_appointments USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: gov_appointments gov_appointments_self; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_appointments_self ON public.gov_appointments USING (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid))) WITH CHECK (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid)));

--
-- Name: gov_appointments gov_appointments_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_appointments_service ON public.gov_appointments USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: gov_notifications; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.gov_notifications ENABLE ROW LEVEL SECURITY;

--
-- Name: gov_notifications gov_notifications_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_notifications_admin ON public.gov_notifications USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: gov_notifications gov_notifications_self; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_notifications_self ON public.gov_notifications USING (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid))) WITH CHECK (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid)));

--
-- Name: gov_notifications gov_notifications_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_notifications_service ON public.gov_notifications USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: gov_payments; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.gov_payments ENABLE ROW LEVEL SECURITY;

--
-- Name: gov_payments gov_payments_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_payments_admin ON public.gov_payments USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: gov_payments gov_payments_self; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_payments_self ON public.gov_payments USING (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid))) WITH CHECK (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid)));

--
-- Name: gov_payments gov_payments_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_payments_service ON public.gov_payments USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: gov_references; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.gov_references ENABLE ROW LEVEL SECURITY;

--
-- Name: gov_references gov_references_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_references_admin ON public.gov_references USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: gov_references gov_references_self; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_references_self ON public.gov_references USING (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid))) WITH CHECK (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid)));

--
-- Name: gov_references gov_references_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY gov_references_service ON public.gov_references USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: organization_memberships org_memberships_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY org_memberships_admin ON public.organization_memberships USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: organization_memberships org_memberships_member_select; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY org_memberships_member_select ON public.organization_memberships FOR SELECT USING (((current_setting('app.user_role'::text, true) = 'user'::text) AND public.app_is_org_member(org_id, (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid)));

--
-- Name: organization_memberships org_memberships_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY org_memberships_service ON public.organization_memberships USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: organization_memberships; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.organization_memberships ENABLE ROW LEVEL SECURITY;

--
-- Name: organizations; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.organizations ENABLE ROW LEVEL SECURITY;

--
-- Name: organizations organizations_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY organizations_admin ON public.organizations USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: organizations organizations_member; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY organizations_member ON public.organizations USING (((current_setting('app.user_role'::text, true) = 'user'::text) AND public.app_is_org_member(id, (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid))) WITH CHECK (((current_setting('app.user_role'::text, true) = 'user'::text) AND public.app_is_org_member(id, (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid)));

--
-- Name: organizations organizations_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY organizations_service ON public.organizations USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: superadmin_accounts sa_acct_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY sa_acct_admin ON public.superadmin_accounts USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: superadmin_accounts sa_acct_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY sa_acct_service ON public.superadmin_accounts USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: security_events; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.security_events ENABLE ROW LEVEL SECURITY;

--
-- Name: security_events security_events_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY security_events_admin ON public.security_events USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: security_events security_events_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY security_events_service ON public.security_events USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: security_events security_events_user_insert; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY security_events_user_insert ON public.security_events FOR INSERT WITH CHECK (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid)));

--
-- Name: superadmin_accounts; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.superadmin_accounts ENABLE ROW LEVEL SECURITY;

--
-- Name: user_recovery_codes urc_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY urc_admin ON public.user_recovery_codes USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: user_recovery_codes urc_self; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY urc_self ON public.user_recovery_codes USING (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid))) WITH CHECK (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid)));

--
-- Name: user_recovery_codes urc_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY urc_service ON public.user_recovery_codes USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: user_integrations; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.user_integrations ENABLE ROW LEVEL SECURITY;

--
-- Name: user_integrations user_integrations_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY user_integrations_admin ON public.user_integrations USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: user_integrations user_integrations_self; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY user_integrations_self ON public.user_integrations USING (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid))) WITH CHECK (((current_setting('app.user_role'::text, true) = 'user'::text) AND (user_id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid)));

--
-- Name: user_integrations user_integrations_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY user_integrations_service ON public.user_integrations USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- Name: user_recovery_codes; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.user_recovery_codes ENABLE ROW LEVEL SECURITY;

--
-- Name: users; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

--
-- Name: users users_admin; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY users_admin ON public.users USING ((current_setting('app.user_role'::text, true) = 'admin'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'admin'::text));

--
-- Name: users users_self; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY users_self ON public.users USING (((current_setting('app.user_role'::text, true) = 'user'::text) AND (id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid))) WITH CHECK (((current_setting('app.user_role'::text, true) = 'user'::text) AND (id = (NULLIF(current_setting('app.user_id'::text, true), ''::text))::uuid)));

--
-- Name: users users_service; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY users_service ON public.users USING ((current_setting('app.user_role'::text, true) = 'service'::text)) WITH CHECK ((current_setting('app.user_role'::text, true) = 'service'::text));

--
-- PostgreSQL database dump complete
--

-- ── app_user privilege tightening (former migration 17_least_privilege_config_grants) ──
-- Defense-in-depth for the GLOBAL config tables (RBAC catalogue + AI prompts /
-- knowledge). These are not per-user tables, so they intentionally carry no
-- Row-Level Security — which means the ONLY DB-level backstop against a missed
-- handler authz check is the app role's table privileges. initdb grants the app
-- role broad SELECT/INSERT/UPDATE/DELETE on every table (it runs before these
-- tables exist, so it cannot be table-specific), so we narrow those grants here
-- to exactly what the repository layer actually uses. After this migration the
-- app connection cannot INSERT a new AI prompt key, rewrite the permission
-- catalogue, or tamper with the knowledge base even if an API authz check is
-- ever bypassed.
--
-- Guarded on the app role being named 'app_user' (the documented default —
-- APP_DB_USER). A deployment that uses a different role name, or an existing DB
-- provisioned without the initdb app role, is left untouched (no-op) and should
-- mirror these REVOKEs by hand for the same backstop. REVOKE of a not-held
-- privilege is a no-op, so re-running is safe.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
        RAISE NOTICE 'app_user role not found — skipping config-table privilege tightening (custom APP_DB_USER? mirror these REVOKEs by hand)';
        RETURN;
    END IF;

    -- permissions: the catalogue is code/migration-defined; the app only reads
    -- it (rbac ListPermissions). No app write path exists.
    REVOKE INSERT, UPDATE, DELETE ON permissions FROM app_user;

    -- role_permissions: rbac replaces grants with DELETE + INSERT (no UPDATE);
    -- both columns form the PK so an UPDATE is meaningless anyway. roles itself
    -- keeps full CRUD — admins manage roles at runtime.
    REVOKE UPDATE ON role_permissions FROM app_user;

    -- ai_prompts: SetPrompt is UPDATE-only against the seeded keys, so the
    -- prompt surface must not grow or shrink through the app. Enforce it in the
    -- DB, not just the repository comment.
    REVOKE INSERT, DELETE ON ai_prompts FROM app_user;

    -- ai_knowledge: the app only runs the search_knowledge SELECT; content is
    -- seed/migration-managed, with no app write path.
    REVOKE INSERT, UPDATE, DELETE ON ai_knowledge FROM app_user;
END $$;
