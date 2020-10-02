package storage

import (
	"fmt"

	"gopkg.in/auth0.v4/management"
)

type LocalJobManager struct {
}

func (ljm *LocalJobManager) VerifyEmail(j *management.Job) error {
	// This is a bit of a hack but there isn't any other way to introduce failure here
	if *j.Status == "FAIL" {
		return fmt.Errorf("Failed to send email")
	}
	return nil
}
