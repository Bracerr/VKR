ALTER TABLE document_types
  DROP COLUMN IF EXISTS reader_roles,
  DROP COLUMN IF EXISTS writer_roles;
