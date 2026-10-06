-- Adults can be registered without a parent; minors are validated by the service.
-- Preserve the foreign key for non-null parent IDs.
BEGIN;
ALTER TABLE public.patients ALTER COLUMN parent_id DROP NOT NULL;
COMMIT;
