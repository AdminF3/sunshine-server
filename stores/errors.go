package stores

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type errorMap map[string]error

func newErrorMap(e *sync.Map) error {
	em := make(errorMap)

	e.Range(func(key interface{}, value interface{}) bool {
		k, kok := key.(string)
		v, vok := value.(error)
		if kok && vok {
			em[k] = v
		}
		return true
	})

	if len(em) == 0 {
		return nil
	}
	return em
}

func (e errorMap) Error() string {
	var b strings.Builder

	for key, value := range e {
		b.WriteString(fmt.Sprintf("%s: %v; ", key, value))
	}

	return b.String()
}

func WithID(err error, id uuid.UUID, kind string) error {
	return fmt.Errorf("%w: %v of type: %v", err, id, kind)
}

func WithIndex(err error, i, v string) error {
	return fmt.Errorf("%w: index %v with value: %v", err, i, v)
}

// IsPQ reports whether given errors comes from github.com/lib/pq.
func IsPQ(err error) bool {
	return errors.As(err, new(*pq.Error))
}

// PQ type assers err to *pq.Error. If given err is not *pq.Error returns a nil pointer.
func PQ(err error) *pq.Error {
	var pqe *pq.Error
	errors.As(err, &pqe)
	return pqe
}

// IsDuplicatedRecord reports whether the error is regarding violated unique
// contraint in the database. Such an error is expected when trying to
// create/update database record.
func IsDuplicatedRecord(err error) bool {
	return IsPQ(err) && PQ(err).Code == "23505"
}

// IsRecordNotFound reports whether the error is regarding searched record which is
// not found in the database.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IsInvalidToken(err error) bool {
	return errors.Is(err, errInvalidToken)
}
