package stores

import (
	"context"
	"sync"

	"stageai.tech/sunshine/sunshine/models"
)

type Stats struct {
	Assets        int64 `json:"assets"`
	Organizations int64 `json:"organizations"`
	Projects      int64 `json:"projects"`
	Users         int64 `json:"users"`
}

type entityStats struct {
	kind    string
	country models.Country
	count   int64
}

func CountryStats(ctx context.Context, s Store, c models.Country) (map[models.Country]Stats, error) {
	entities := []models.Entity{
		new(models.Asset),
		new(models.Organization),
		new(models.Project),
		new(models.User),
	}
	ch := make(chan entityStats)
	result := make(map[models.Country]Stats, len(entities))
	errs := new(sync.Map)

	var wg sync.WaitGroup
	wg.Add(len(entities))
	go func() { wg.Wait(); close(ch) }()
	for _, e := range entities {
		go func(e models.Entity) {
			defer wg.Done()
			db := s.DB().Table(e.TableName()).
				Select("count(id), country").
				Group("country")

			if c != "" {
				db = db.Where("country = ?", c)
			}

			rows, err := db.Rows()
			if err != nil {
				errs.Store(e.Kind(), err)
				return
			}

			var (
				country models.Country
				count   int64
			)
			defer rows.Close()
			for rows.Next() {
				if err := rows.Scan(&count, &country); err != nil {
					errs.Store(e.Kind()+"_scan", err)
					return
				}
				ch <- entityStats{
					kind:    e.Kind(),
					country: country,
					count:   count,
				}
			}
		}(e)
	}

	for v := range ch {
		cs := result[v.country]
		switch v.kind {
		case "asset":
			cs.Assets = v.count
		case "organization":
			cs.Organizations = v.count
		case "project":
			cs.Projects = v.count
		case "user":
			cs.Users = v.count
		}
		result[v.country] = cs
	}

	return result, newErrorMap(errs)
}
