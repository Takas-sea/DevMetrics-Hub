DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'devmetrics_user') THEN
    CREATE ROLE devmetrics_user LOGIN PASSWORD 'devmetrics_password';
  END IF;
END
$$;

ALTER DATABASE devmetrics_hub OWNER TO devmetrics_user;
GRANT ALL PRIVILEGES ON DATABASE devmetrics_hub TO devmetrics_user;