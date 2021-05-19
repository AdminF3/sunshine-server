package models

import (
	"database/sql/driver"
	"fmt"
	"net/url"

	"github.com/google/uuid"
)

// Attachment represents a file attachment regardless of where is stored.
type Attachment struct {
	Value

	// Owner holds ID of the entity this attachment is associated with.
	Owner uuid.UUID `gorm:"column:owner_id"`

	// Name of the attachment.
	Name string `json:"name"`

	// ContentType is the MIME type of the attachment contents.
	ContentType string `json:"content_type"`

	// UploadType is the milestone types.
	UploadType UploadType `json:"upload_type" validate:"omitempty,upload_type"`

	// Size records the uncompressed size of the attachment. The value -1
	// indicates that the length is unknown. Values >= 0 indicate that the
	// given number of bytes may be read from Content.
	Size int64 `json:"length"`

	Comment string `json:"comment"`
}

func (Attachment) TableName() string {
	return "attachments"
}

func (a *Attachment) AfterFind() {
	a.Name = url.PathEscape(a.Name)

}

type UploadType string

var UploadTypes = map[string]struct{}{
	"general leaflet":                                              struct{}{},
	"process leaflet":                                              struct{}{},
	"aquisition protocol meeting":                                  struct{}{},
	"aquisition protocol survey":                                   struct{}{},
	"energy audit report":                                          struct{}{},
	"technical inspection report":                                  struct{}{},
	"esco tender":                                                  struct{}{},
	"commitment protocol meeting":                                  struct{}{},
	"commitment protocol survey":                                   struct{}{},
	"pre epc agreement":                                            struct{}{},
	"cooperation agreement":                                        struct{}{},
	"construction project":                                         struct{}{},
	"procurement construction installation":                        struct{}{},
	"finincing offer and altum":                                    struct{}{},
	"draft epc contract":                                           struct{}{},
	"kickoff protocol meeting":                                     struct{}{},
	"kickoff protocol survey":                                      struct{}{},
	"signed epc":                                                   struct{}{},
	"agreement altum bank loan":                                    struct{}{},
	"agreement altum grand agreement":                              struct{}{},
	"project management contract":                                  struct{}{},
	"construction works company contract":                          struct{}{},
	"engineering networks company contract":                        struct{}{},
	"supervision contract":                                         struct{}{},
	"maintenance company contract":                                 struct{}{},
	"land owners contract":                                         struct{}{},
	"house heating contract":                                       struct{}{},
	"windows contract":                                             struct{}{},
	"kick-off meeting document":                                    struct{}{},
	"initial information meeting document":                         struct{}{},
	"weekly meeting report document":                               struct{}{},
	"informative residents meeting document":                       struct{}{},
	"other residents meeting document":                             struct{}{},
	"monthly construction company report":                          struct{}{},
	"expenses document":                                            struct{}{},
	"tama comments document":                                       struct{}{},
	"building audit document":                                      struct{}{},
	"work acceptance document":                                     struct{}{},
	"construction managers final meeting mom":                      struct{}{},
	"residents final meeting mom":                                  struct{}{},
	"building user guide":                                          struct{}{},
	"fa financial statements":                                      struct{}{},
	"fa bank confirmation":                                         struct{}{},
	"commissioning report":                                         struct{}{},
	"insurance policies":                                           struct{}{},
	"independent energy audit measurement and verification report": struct{}{},
	"defects declarations":                                         struct{}{},
	"letter of information":                                        struct{}{},
	"payment of loans":                                             struct{}{},
	"investment invoices":                                          struct{}{},
	"annual check other financials":                                struct{}{},
	"building maintenance milestone":                               struct{}{},
	"building inspection document":                                 struct{}{},
	"acquisition other":                                            struct{}{},
	"feasibility other":                                            struct{}{},
	"commitment other":                                             struct{}{},
	"design other":                                                 struct{}{},
	"preparation other":                                            struct{}{},
	"kickoff other":                                                struct{}{},
	"signed other":                                                 struct{}{},
	"renovation other":                                             struct{}{},
	"commissioning other":                                          struct{}{},
	"technical other":                                              struct{}{},
	"forfaiting other":                                             struct{}{},
	"results other":                                                struct{}{},
	"payment other":                                                struct{}{},
	"annual other":                                                 struct{}{},
	"building other":                                               struct{}{},
	"registration document":                                        struct{}{},
	"proof of address":                                             struct{}{},
	"vat document":                                                 struct{}{},
	"energy management system company":                             struct{}{},
	"forfaiting annual check":                                      struct{}{},
	"lear apply":                                                   struct{}{},
	"proof of transfer":                                            struct{}{},
	"epc contracts":                                                struct{}{},
}

// Scan implements the database/sql.Scanner interface.
func (t *UploadType) Scan(value interface{}) error {
	var v, ok = value.([]byte)
	if !ok {
		return fmt.Errorf("invalid upload type: %v", v)
	}

	*t = UploadType(v)
	return nil
}

// Value implements the database/sql/driver.Valuer interface.
func (t UploadType) Value() (driver.Value, error) {
	if len(t) == 0 {
		return nil, nil
	}
	return string(t), nil
}
