package service

import (
	cloud "github.com/terramate-io/terramate/cloud"
)

type ErrMsg error
type UserMsg struct {
	User cloud.User
}
