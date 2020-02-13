package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/robfig/cron/v3"
)

func validateJSON(val interface{}) error {
	if str, ok := val.(string); ok {
		valid := json.Valid([]byte(str))
		if valid {
			return nil
		}
		return fmt.Errorf("value `%s` is not a valid JSON", str)
	}
	return errors.New("value must be string")
}

func validateCron(val interface{}) error {
	if str, ok := val.(string); ok {
		_, err := cron.ParseStandard(str)
		if err == nil {
			return nil
		}
		return fmt.Errorf("value `%s` is not a valid cron expression: %v", str, err)
	}
	return errors.New("value must be string")

}
