package project_intake

import "time"

type ProjectAppliedEvent struct {
	ID             int64
	CompanyName    string
	ProjectScope   string
	Budget         int64
	ContactEmail   string
	ContactInfo    string
	DeadlineNeeded bool
	DeadlineSpec   string
	IsCustomBudget bool
	Timestamp      time.Time
}
