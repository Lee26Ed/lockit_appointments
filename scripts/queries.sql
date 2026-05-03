SELECT id, payload, created_at
FROM jobs
ORDER BY created_at DESC
LIMIT 10;

SELECT id, 
payload->>'original_filename' AS filename
FROM jobs
WHERE payload->>'mime_type' = 'image/jpeg';

  SELECT 
    public_id,
    status,
    progress,
    payload->>'original_filename' AS filename
FROM jobs;

SELECT 
    public_id,
    status,
    progress
FROM jobs
WHERE public_id = '000e327e-616a-46fd-a5f0-71930b4ba199';

SELECT id, payload
FROM   jobs
WHERE  status = 'pending'
ORDER  BY created_at
LIMIT  1;

SELECT * 
FROM jobs 
LIMIT 5;

SELECT
    public_id,
    status,
    created_at,
    updated_at
FROM jobs
ORDER BY created_at;

-- List the indexes
SELECT indexname, indexdef
FROM   pg_indexes
WHERE  tablename = 'jobs'
ORDER  BY indexname;
