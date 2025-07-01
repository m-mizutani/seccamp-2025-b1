-- 過去24時間で複数の異なるIPアドレスから同一ユーザーへの認証失敗を検知
SELECT 
    user,
    COUNT(DISTINCT remote) as unique_ips,
    COUNT(*) as total_failures,
    array_agg(DISTINCT remote) as source_ips,
    MIN(from_unixtime(timestamp / 1000)) as first_failure,
    MAX(from_unixtime(timestamp / 1000)) as last_failure
FROM 
    service_logs
WHERE 
    action = 'login'
    AND success = false
    AND timestamp > (unix_timestamp() - 86400) * 1000  -- 過去24時間
GROUP BY 
    user
HAVING 
    COUNT(DISTINCT remote) >= 3
    AND COUNT(*) >= 10
ORDER BY 
    unique_ips DESC, total_failures DESC 