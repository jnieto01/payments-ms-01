package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringSlice is a JSON-serializable []string for GORM
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringSlice) Scan(value interface{}) error {
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into StringSlice", value)
	}
	return json.Unmarshal(raw, s)
}

type Plan struct {
	ID                   uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	PlanKey              string      `gorm:"type:varchar(50);not null;uniqueIndex" json:"plan_key"`
	Name                 string      `gorm:"type:varchar(100);not null" json:"name"`
	ShortName            string      `gorm:"type:varchar(20);not null;default:''" json:"short_name"`
	Price                float64     `gorm:"type:decimal(12,2);not null" json:"price"`
	Currency             string      `gorm:"type:varchar(10);not null;default:ARS" json:"currency"`
	Period               string      `gorm:"type:varchar(50);not null;default:monthly" json:"period"`
	Features             StringSlice `gorm:"type:json;not null" json:"features"`
	TrialDays            int         `gorm:"not null;default:14" json:"trial_days"`
	WarningDays          int         `gorm:"not null;default:5" json:"warning_days"`
	IsActive             bool        `gorm:"not null;default:true" json:"is_active"`
	MPPreapprovalPlanID  string      `gorm:"type:varchar(255)" json:"mp_preapproval_plan_id,omitempty"`
	MPPreapprovalPlanURL string      `gorm:"type:text" json:"mp_preapproval_plan_url,omitempty"`
	CreatedAt            time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
}
