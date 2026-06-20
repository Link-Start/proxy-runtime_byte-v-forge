package sqlite

const (
	sqliteLeaseActiveUntilPredicate    = `(expires_at='' OR expires_at>?)`
	sqliteLeaseExpiredByPredicate      = `(expires_at!='' AND expires_at<=?)`
	sqliteCleanupPendingLeasePredicate = `(
  json_extract(lease_json, '$.session.labels.route_cleanup_pending')='true'
  OR json_extract(lease_json, '$.session.labels.provider_cleanup_pending')='true'
)`
)
