package auditlog

const (
	SnapshotPending         = "pending"
	SnapshotRecorded        = "recorded"
	SnapshotUnchanged       = "unchanged"
	SnapshotNotApplicable   = "not_applicable"
	SnapshotNotConfigured   = "not_configured"
	SnapshotOperationFailed = "operation_failed"
	SnapshotUnavailable     = "unavailable"
	SnapshotLegacyUnknown   = "legacy_unknown"
)

func ValidSnapshotStatus(status string) bool {
	switch status {
	case SnapshotPending, SnapshotRecorded, SnapshotUnchanged, SnapshotNotApplicable, SnapshotNotConfigured, SnapshotOperationFailed, SnapshotUnavailable, SnapshotLegacyUnknown:
		return true
	default:
		return false
	}
}
