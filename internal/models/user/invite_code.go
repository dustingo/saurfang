package user

// InviteCodes 注册邀请码
type InviteCodes struct {
	Id   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Code string `gorm:"type:varchar(100);default:null" json:"code"`
	Used uint   `gorm:"type:int;default:0;comment:是否使用" json:"used"`
}
