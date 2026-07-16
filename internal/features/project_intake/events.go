package project_intake

import "time"

type ProjectAppliedEvent struct {
	ID             int64     `json:"id"`
	CompanyName    string    `json:"companyName"`
	ProjectScope   string    `json:"projectScope"`
	Budget         int64     `json:"budget"`
	ContactEmail   string    `json:"contactEmail"`
	ContactInfo    string    `json:"contactInfo"`
	DeadlineNeeded bool      `json:"deadlineNeeded"`
	DeadlineSpec   string    `json:"deadlineSpec"`
	IsCustomBudget bool      `json:"isCustomBudget"`
	Timestamp      time.Time `json:"timestamp"`
}
