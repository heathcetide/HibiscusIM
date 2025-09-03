package models

import "time"

// SOS Alert（求助警报）
type Alert struct {
	ID           uint   `gorm:"primaryKey"`
	UserID       uint   // 触发者的用户ID
	AlertType    string // "SOS"
	Status       string // "pending" "processing" "completed" "cancelled"
	Priority     string // "high" (紧急)
	AlertDetails string // 求助信息，JSON 格式：姓名、电话、地址、位置坐标
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// 用户执行的操作（回拨、拨打急救等）
type AlertAction struct {
	ID         uint   `gorm:"primaryKey"`
	AlertID    uint   // 对应的 SOS 警报 ID
	Action     string // "call_120", "call_back", "safe"
	ActionTime time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
