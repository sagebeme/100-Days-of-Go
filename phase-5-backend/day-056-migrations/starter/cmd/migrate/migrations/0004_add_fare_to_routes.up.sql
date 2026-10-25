-- The fare in whole shillings. Existing routes get Ksh100 until someone sets theirs.
ALTER TABLE routes ADD COLUMN fare integer NOT NULL DEFAULT 100 CHECK (fare >= 0);
