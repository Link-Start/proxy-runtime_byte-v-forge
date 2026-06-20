package postgres

const (
	postgresLeaseActiveUntilNowPredicate = `(expires_at IS NULL OR expires_at > now())`
	postgresLeaseExpiredByNowPredicate   = `(expires_at IS NOT NULL AND expires_at <= now())`
	postgresLeaseCleanupPendingPredicate = `(
		lease_json #>> '{session,labels,route_cleanup_pending}' = 'true'
		OR lease_json #>> '{session,labels,provider_cleanup_pending}' = 'true'
	)`
)
