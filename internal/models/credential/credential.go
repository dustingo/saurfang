package credential

// AKSKCredential AK/SK认证结构
type UserCredential struct {
	ID        uint   `json:"id"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	UserID    uint   `gorm:"unique" json:"user_id"` // 关联的用户ID
	Status    string `json:"status"`                // active, inactive
}
type AKSKAuthService struct {
	Credentials map[string]*UserCredential `json:"credentials"`
	JWTSecret   string                     `json:"jwt_secret"`
}
