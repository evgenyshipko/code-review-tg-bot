package vacation

// состояние админ-панели
type AdminPanelState struct {
	State       string `json:"state"`
	SelectedUID int64  `json:"selected_uid,omitempty"`
}
