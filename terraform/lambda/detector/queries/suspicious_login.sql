-- 過去1時間で同一IPアドレスから5回以上のログイン失敗を検知
SELECT 
    remote,
    user,
    COUNT(*) as failed_attempts,
    MIN(from_unixtime(timestamp / 1000)) as first_attempt,
    MAX(from_unixtime(timestamp / 1000)) as last_attempt
FROM 
    service_logs
WHERE 
    action = 'login'
    AND success = false
    AND timestamp > (unix_timestamp() - 3600) * 1000  -- 過去1時間
GROUP BY 
    remote, user
HAVING 
    COUNT(*) >= 5
ORDER BY 
    failed_attempts DESC 