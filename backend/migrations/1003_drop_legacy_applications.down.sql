-- Хуучин overlay хүснэгтүүдийг сэргээнэ (бүтэц; өгөгдөл нь oauth_clients-д).

CREATE TABLE IF NOT EXISTS public.applications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    client_id text NOT NULL,
    name text NOT NULL,
    app_type text DEFAULT 'm2m'::text NOT NULL,
    tags text[] DEFAULT '{}'::text[] NOT NULL,
    redirect_uris text[] DEFAULT '{}'::text[] NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    created_by text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone,
    CONSTRAINT applications_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.application_services (
    application_id uuid NOT NULL,
    service_id uuid NOT NULL
);
