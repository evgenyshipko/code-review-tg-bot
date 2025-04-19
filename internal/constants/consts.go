package constants

import "code-review-tg-bot/internal/access"

const (
	Start           = "start"
	Rest            = "rest"
	Work            = "work"
	Vacations_start = "vacations_start"
	Vacations       = "vacations"
	Test_vacation   = "test_vacation"
)

var CommandToRoleMapping = map[string][]access.UserRole{
	Start:           []access.UserRole{},
	Rest:            []access.UserRole{access.Reviewer},
	Work:            []access.UserRole{access.Reviewer},
	Vacations:       []access.UserRole{access.Admin},
	Vacations_start: []access.UserRole{access.Admin},
	Test_vacation:   []access.UserRole{access.Tester, access.Admin},
}
