-- 過去1時間で同一ユーザーから100回以上のread操作を検知
SELECT 
    user,
    remote,
    COUNT(*) as read_count,
    COUNT(DISTINCT target) as unique_targets,
    MIN(from_unixtime(timestamp / 1000)) as first_access,
    MAX(from_unixtime(timestamp / 1000)) as last_access
FROM 
    service_logs
WHERE 
    action = 'read'
    AND success = true
    AND timestamp > (unix_timestamp() - 3600) * 1000  -- 過去1時間
GROUP BY 
    user, remote
HAVING 
    COUNT(*) >= 100
ORDER BY 
    read_count DESC 