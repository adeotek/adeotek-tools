package models

import "time"

// ScriptExecutionHistory represents a record of an executed migration script
type ScriptExecutionHistory struct {
	ScriptFile string    `db:"ScriptFile"`
	ScriptHash string    `db:"ScriptHash"`
	ExecutedAt time.Time `db:"ExecutedAt"`
}
