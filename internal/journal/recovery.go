package journal

// ObservedFile is the current vault state for a path. Hash is the SHA-256
// digest of its current bytes when Exists is true.
type ObservedFile struct {
	Exists bool
	Hash   string
}

// RecoveryState classifies an observed path against the state recorded before
// and after an operation. Unknown states are conflicts and are never written.
type RecoveryState string

const (
	RecoveryAtBefore RecoveryState = "before"
	RecoveryAtAfter  RecoveryState = "after"
	RecoveryConflict RecoveryState = "conflict"
)

func Classify(action FileAction, current ObservedFile) RecoveryState {
	if current.Exists == action.AfterExists && (!current.Exists || current.Hash == action.AfterHash) {
		return RecoveryAtAfter
	}
	if current.Exists == action.BeforeExists && (!current.Exists || current.Hash == action.BeforeHash) {
		return RecoveryAtBefore
	}
	return RecoveryConflict
}
