package models

import "time"

type RecordingPrompt struct {
	ID        uint   `gorm:"primaryKey"`
	Text      string `gorm:"size:1024"` // 屏幕上显示的待读文本
	Order     int    // 第几句，从1开始
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Recording struct {
	ID            uint   `gorm:"primaryKey"`
	UserID        uint   // 录音者
	PromptID      uint   // 对应哪一句录音
	SentenceIndex int    // 句子编号（冗余）
	FileURL       string `gorm:"size:1024"` // 存储到对象存储后的 URL
	Format        string `gorm:"size:32"`   // e.g. "wav", "opus"
	DurationMs    int    // 毫秒
	SizeBytes     int64
	Checksum      string `gorm:"size:128"`
	Status        string `gorm:"size:32"`   // uploaded / processing / ready / failed
	Transcription string `gorm:"type:text"` // 可选：自动语音识别结果
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type VoiceJob struct {
	ID           uint `gorm:"primaryKey"`
	UserID       uint
	RecordingIDs string `gorm:"type:text"` // JSON array of recording IDs
	Status       string `gorm:"size:32"`   // pending/processing/succeeded/failed
	ResultURL    string `gorm:"size:1024"` // 生成结果（模型或合成音）存放地址
	Progress     int    // 0-100
	ErrorMessage string `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
