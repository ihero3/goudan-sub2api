DO $$
DECLARE
    dt date;
BEGIN
    FOR dt IN SELECT DISTINCT DATE(ul.created_at) FROM usage_logs ul JOIN api_keys ak ON ak.id = ul.api_key_id WHERE ak.team_id = 1 AND ul.created_at >= '2026-07-27' LOOP
        -- Team daily
        INSERT INTO team_usage_team_daily (team_id, bucket_date, total_requests, input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens, total_cost, actual_cost, computed_at, created_at, updated_at)
        SELECT 1, dt, COUNT(*), COALESCE(SUM(ul.input_tokens),0), COALESCE(SUM(ul.output_tokens),0), 0, 0, COALESCE(SUM(ul.total_cost),0), COALESCE(SUM(ul.actual_cost),0), NOW(), NOW(), NOW()
        FROM usage_logs ul JOIN api_keys ak ON ak.id = ul.api_key_id
        WHERE ak.team_id = 1 AND ul.created_at >= dt AND ul.created_at < (dt + INTERVAL '1 day')
        ON CONFLICT (team_id, bucket_date) DO UPDATE SET total_requests = EXCLUDED.total_requests, input_tokens = EXCLUDED.input_tokens, output_tokens = EXCLUDED.output_tokens, total_cost = EXCLUDED.total_cost, actual_cost = EXCLUDED.actual_cost, computed_at = EXCLUDED.computed_at, updated_at = EXCLUDED.updated_at;

        -- Dept daily
        INSERT INTO team_usage_dept_daily (team_id, department_id, department_name, cost_center_code, bucket_date, total_requests, input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens, total_cost, actual_cost, computed_at, created_at, updated_at)
        SELECT 1, ak.department_id, MAX(d.name), MAX(d.cost_center_code), dt, COUNT(*), COALESCE(SUM(ul.input_tokens),0), COALESCE(SUM(ul.output_tokens),0), 0, 0, COALESCE(SUM(ul.total_cost),0), COALESCE(SUM(ul.actual_cost),0), NOW(), NOW(), NOW()
        FROM usage_logs ul JOIN api_keys ak ON ak.id = ul.api_key_id LEFT JOIN departments d ON d.id = ak.department_id
        WHERE ak.team_id = 1 AND ak.department_id IS NOT NULL AND ul.created_at >= dt AND ul.created_at < (dt + INTERVAL '1 day')
        GROUP BY ak.department_id
        ON CONFLICT (team_id, department_id, bucket_date) DO UPDATE SET department_name = EXCLUDED.department_name, total_requests = EXCLUDED.total_requests, input_tokens = EXCLUDED.input_tokens, output_tokens = EXCLUDED.output_tokens, total_cost = EXCLUDED.total_cost, actual_cost = EXCLUDED.actual_cost, computed_at = EXCLUDED.computed_at, updated_at = EXCLUDED.updated_at;

        -- Consumer daily
        INSERT INTO team_usage_consumer_daily (team_id, consumer_id, consumer_name, consumer_type, bucket_date, total_requests, input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens, total_cost, actual_cost, computed_at, created_at, updated_at)
        SELECT 1, ak.consumer_id, MAX(c.name), MAX(c.type), dt, COUNT(*), COALESCE(SUM(ul.input_tokens),0), COALESCE(SUM(ul.output_tokens),0), 0, 0, COALESCE(SUM(ul.total_cost),0), COALESCE(SUM(ul.actual_cost),0), NOW(), NOW(), NOW()
        FROM usage_logs ul JOIN api_keys ak ON ak.id = ul.api_key_id LEFT JOIN consumers c ON c.id = ak.consumer_id
        WHERE ak.team_id = 1 AND ak.consumer_id IS NOT NULL AND ul.created_at >= dt AND ul.created_at < (dt + INTERVAL '1 day')
        GROUP BY ak.consumer_id
        ON CONFLICT (team_id, consumer_id, bucket_date) DO UPDATE SET consumer_name = EXCLUDED.consumer_name, consumer_type = EXCLUDED.consumer_type, total_requests = EXCLUDED.total_requests, input_tokens = EXCLUDED.input_tokens, output_tokens = EXCLUDED.output_tokens, total_cost = EXCLUDED.total_cost, actual_cost = EXCLUDED.actual_cost, computed_at = EXCLUDED.computed_at, updated_at = EXCLUDED.updated_at;

        -- Model daily
        INSERT INTO team_usage_model_daily (team_id, department_id, consumer_id, bucket_date, model_name, total_requests, input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens, total_cost, actual_cost, computed_at, created_at, updated_at)
        SELECT 1, ak.department_id, ak.consumer_id, dt, COALESCE(NULLIF(ul.model,''),'unknown'), COUNT(*), COALESCE(SUM(ul.input_tokens),0), COALESCE(SUM(ul.output_tokens),0), 0, 0, COALESCE(SUM(ul.total_cost),0), COALESCE(SUM(ul.actual_cost),0), NOW(), NOW(), NOW()
        FROM usage_logs ul JOIN api_keys ak ON ak.id = ul.api_key_id
        WHERE ak.team_id = 1 AND ak.department_id IS NOT NULL AND ak.consumer_id IS NOT NULL AND ul.created_at >= dt AND ul.created_at < (dt + INTERVAL '1 day')
        GROUP BY ak.department_id, ak.consumer_id, COALESCE(NULLIF(ul.model,''),'unknown')
        ON CONFLICT (team_id, department_id, consumer_id, bucket_date, model_name) DO UPDATE SET total_requests = EXCLUDED.total_requests, input_tokens = EXCLUDED.input_tokens, output_tokens = EXCLUDED.output_tokens, total_cost = EXCLUDED.total_cost, actual_cost = EXCLUDED.actual_cost, computed_at = EXCLUDED.computed_at, updated_at = EXCLUDED.updated_at;
    END LOOP;
END $$;
