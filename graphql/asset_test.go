package graphql

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"
)

func TestAssetListing(t *testing.T) {
	e := services.NewTestEnv(t)

	u := stores.NewTestUser(t, e.UserStore)
	pd := stores.NewTestAdminNwManager(t, e.UserStore)

	makeTestAssets(t, e, 9, uuid.Nil)
	makeTestAssets(t, e, 1, u.ID)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		args   string
		result string
		query  string
	}{
		{
			name:   "unauth",
			ctx:    ctx,
			errors: []string{controller.ErrUnauthorized.Error()},
			query:  "query_assetListing_request.json",
		},
		{
			name:   "all",
			ctx:    services.NewTestContext(t, e, u),
			result: LoadGQLTestFile(t, "query_assetListing_response.json", strMul(10), 10),
			query:  "query_assetListing_request.json",
		},
		{
			name:   "all-reports",
			ctx:    services.NewTestContext(t, e, pd),
			result: LoadGQLTestFile(t, "query_assetReports_response.json", strMul(10), 10),
			query:  "query_assetReports_request.json",
		},
		{
			name:   "all-reports filterMine",
			ctx:    services.NewTestContext(t, e, u),
			result: LoadGQLTestFile(t, "query_assetListing_response.json", strMul(1), 1),
			args:   `filterMine: true`,
			query:  "query_assetListing_request.json",
		},
		{
			name:   "first 5",
			ctx:    services.NewTestContext(t, e, u),
			result: LoadGQLTestFile(t, "query_assetListing_response.json", strMul(5), 10),
			args:   `first: 5, offset: 0`,
			query:  "query_assetListing_request.json",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   LoadGQLTestFile(t, c.query, c.args),
				Result:  c.result,
			})
		})
	}
}

func strMul(c int) string {
	str := `{"address": "End of the world","ownerName": "Goo Corporation","residentsCount": 1}`
	rows := strings.Repeat(fmt.Sprintf("%[1]s ,", str), c)
	return rows[:len(rows)-1]
}

func makeTestAssets(t *testing.T, e *services.Env, count int, userID uuid.UUID) []models.Asset {
	var ownerID uuid.UUID
	if userID != uuid.Nil {
		ownerID = stores.NewTestOrg(t, e.OrganizationStore, userID).ID
	}

	assets := make([]models.Asset, count)
	for i := range assets {
		asset := stores.NewTestAsset(t, e.AssetStore, func(_ *testing.T, _ stores.Store, d *models.Document) {
			if ownerID != uuid.Nil {
				d.Data.(*models.Asset).Owner = ownerID
			}
		})

		assets[i] = *asset.Data.(*models.Asset)
	}

	return assets
}
