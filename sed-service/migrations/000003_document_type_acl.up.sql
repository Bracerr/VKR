ALTER TABLE document_types
  ADD COLUMN reader_roles TEXT[] NOT NULL DEFAULT '{}',
  ADD COLUMN writer_roles TEXT[] NOT NULL DEFAULT '{}';
