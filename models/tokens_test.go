package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"
)

func TestValidToken(t *testing.T) {
	var (
		validate = validator.New()

		tt = []struct {
			purpose TokenPurpose
			now     time.Duration
			valid   bool
		}{
			{SessionToken, 20 * timeDay, true},
			{SessionToken, 31 * timeDay, false},
			{CreateToken, 10 * time.Hour, true},
			{CreateToken, 50 * time.Hour, false},
			{ResetPwdToken, 10 * time.Hour, true},
			{ResetPwdToken, 25 * time.Hour, false},
			{"world", 1 * time.Second, false},
		}
	)
	defer func() {
		now = time.Now
	}()

	for _, tc := range tt {
		token := NewToken(tc.purpose, uuid.New())

		t.Run(fmt.Sprintf("expire %s %s", tc.purpose, tc.now), func(t *testing.T) {
			now = func() time.Time {
				return token.CreatedAt.Add(tc.now)
			}
			result := ValidToken(*token, tc.purpose, validate)
			if tc.valid != result {
				t.Logf("token.CreatedAt = %+v\n", token.CreatedAt)
				t.Logf("now() = %+v\n", now())
				t.Logf("token.TTL = %+v\n", token.TTL)
				t.Errorf("%v expected valid to be %t got %t",
					token, tc.valid, result)
			}
		})

		t.Run(fmt.Sprintf("bad resource %s %s", tc.purpose, tc.now), func(t *testing.T) {
			if ValidToken(*token, "world", validate) {
				t.Error("Should be invalid when tested against different resource")
			}
		})

		t.Run(fmt.Sprintf("no entity %s %s", tc.purpose, tc.now), func(t *testing.T) {
			token.UserID = uuid.Nil
			if ValidToken(*token, tc.purpose, validate) {
				t.Error("Should be invalid with bad entity")
			}
		})
	}
}
