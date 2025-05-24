DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relname = 'users_email_unique'
          AND n.nspname = 'public'
    ) THEN
        CREATE UNIQUE INDEX users_email_unique ON public.users (email);
    END IF;
END
$$;
