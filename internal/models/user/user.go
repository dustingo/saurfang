package user

// // User 用户
//
//	type User struct {
//		gorm.Model
//		Username string `gorm:"unique;not null" form:"username"`
//		Password string `gorm:"not null" form:"password"`
//		Token    string `gorm:"text" json:"token"`
//		Code     string `gorm:"unique;type:varchar(100);not null" json:"code"`
//		Roles    []Role `gorm:"many2many:user_roles;"`
//	}
//
// RegisterPayload 注册
type RegisterPayload struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

// UserInfo 用户简略信息
type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	RoleID   int    `json:"role_id"`
	Name     string `json:"name"` // role别名
}
type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// User 用户
type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"unique" json:"username"`
	Password string `json:"-"`
	Token    string `json:"token"`
	Code     string `json:"code"`
	Roles    Role   `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// Role 角色
type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"unique" json:"name"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}
type RolePayload struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique" json:"name"`
}

// UserRole 用户和角色映射
type UserRole struct {
	RoleID uint `json:"role_id"`
	UserID uint `json:"user_id"`
}

// Permission 路由(组)记录
type Permission struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `gorm:"index" json:"name"`
	Group string `gorm:"index" json:"group"`
}
type RolePermissionRelation struct {
	RoleID       uint `gorm:"column:role_id"`
	PermissionID uint `gorm:"column:permission_id"`
}
