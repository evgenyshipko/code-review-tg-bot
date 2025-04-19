package vacation

// состояние админ-панели
type AdminPanelState struct {
	State  string `json:"state"`
	UserId int64  `json:"user_id,omitempty"`
}
