package openapi

import "testing"

// TestJSONTag checks if in string there is json tag and extract it
func TestFieldTag(t *testing.T) {
	var cases = []struct {
		name  string
		value string
		tag   string
		fType string
	}{
		{
			name: "ok",
			value: `	Name      string \x60json:"name" validate:"required"\x60 `,
			tag:   "name",
			fType: "string",
		},
		{
			name: "uuid",
			value: `	Owner        uuid.UUID        \x60json:"owner" gorm:"column:owner_id" validate:"required"\x60`,
			tag:   "owner",
			fType: "uuid.UUID",
		},
		{
			name: "array",
			value: `	SocialProfiles    []SocialProfile    \x60json:"social_profiles" gorm:"foreignkey:UserID"\x60`,
			tag:   "social_profiles",
			fType: "[]SocialProfile",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tag := fieldTag(c.value)
			fType := fieldType(c.value)

			if fType != c.fType {
				t.Fatalf("expeted to find %q, found %q for type", c.fType, fType)
			}

			if tag != c.tag {
				t.Fatalf("expeted to find %q, found %q for tag name", c.tag, tag)
			}
		})
	}
}

// TestModelTags checks if in correct model will extracts its tags.
func TestModelTags(t *testing.T) {
	model := parseSrcModel("stageai.tech/sunshine/sunshine/models.User")

	if model.Name != "User" {
		t.Fatalf("expected name %q but got %q", "user", model.Name)
	}

	if len(model.fields) == 0 ||
		model.fields["name"] != "string" {
		t.Fatalf("fields are not imported properly %v", model.fields)
	}

}
