package user

type Response struct {
	Data    ResponseData `json:"data"`
	Message string       `json:"message"`
	Status  int          `json:"status"`
}

type ResponseData struct {
	Items []RoleResponse `json:"items"`
}

type RoleResponse struct {
	ID          uint             `json:"id"`
	Name        string           `json:"name"`
	Permissions []PermissionItem `json:"permissions"`
}

type PermissionItem struct {
	ID    uint   `json:"id"`
	Group string `json:"group"`
}
