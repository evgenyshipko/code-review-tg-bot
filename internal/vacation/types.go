package vacation

import "time"

// статус отпуска пользователя
type StatusVacationUser struct {
	IsOnVacation bool      `json:"is_on_vacation"`
	ReturnDate   time.Time `json:"return_date,omitempty"`
}

// состояние админ-панели
type AdminPanelState struct {
	State       string `json:"state"`
	SelectedUID int64  `json:"selected_uid,omitempty"`
}
