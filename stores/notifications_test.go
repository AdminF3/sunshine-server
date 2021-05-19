package stores

import (
	"testing"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

func TestNotifier(t *testing.T) {
	db := models.NewTestGORM(t)

	notifier := NewNotifier(db, validate)
	st := NewUserStore(db, validate)
	usr := NewTestUser(t, st)
	n := models.Notification{Old: "Heya", RecipientID: usr.ID, Action: models.UserActionCreate, Country: models.CountryBulgaria}
	if err := notifier.Notify(ctx, &n); err != nil {
		t.Fatalf("failed to notify: %s", err)
	}

	if n.ID == uuid.Nil {
		t.Errorf("expected a notification to have been created, got: %#v", n)
	}

	t.Run("list", func(t *testing.T) {
		tNotifyList(t, notifier, uuid.New(), "", 0)
		tNotifyList(t, notifier, usr.ID, "", 1)
		tNotifyList(t, notifier, usr.ID, models.UserActionCreate, 1)
		tNotifyList(t, notifier, usr.ID, models.UserActionUpdate, 0)
	})

	t.Run("see", func(t *testing.T) {
		tNotifySee(t, notifier, n)
	})

	t.Run("filter and count", tNotifyFilterCount)

	t.Run("broadcast", func(t *testing.T) {
		NewTestPortfolioRole(t, st, models.PortfolioDirectorRole, models.CountryLatvia)
		NewTestPortfolioRole(t, st, models.CountryAdminRole, models.CountryLatvia)
		pm := NewTestUser(t, st).Data.(*models.User)
		prj := NewTestProject(t, st, TPrjWithPm(pm.ID))
		notifier.Broadcast(ctx, models.UserActionUpload, *pm, *prj, "", "foo.jpg", usr.ID, nil)
		tNotifyList(t, notifier, pm.ID, models.UserActionUpload, 1)
	})
}

func tNotifyFilterCount(t *testing.T) {
	db := models.NewTestGORM(t)

	n := NewNotifier(db, validate)
	a := NewTestAsset(t, NewUserStore(db, validate))
	u := NewTestUser(t, NewUserStore(db, validate))
	u2 := NewTestUser(t, NewUserStore(db, validate))

	NewTestNotification(t, n, u.ID)
	NewTestNotification(t, n, u.ID)
	NewTestNotification(t, n, u.ID)
	NewTestNotification(t, n, u.ID)
	NewTestNotification(t, n, u.ID)
	NewTestNotification(t, n, u.ID)
	NewTestNotification(t, n, u.ID)
	NewTestNotification(t, n, u.ID, TNWithTarget(a))
	NewTestNotification(t, n, u.ID, TNWithAction(models.UserActionCreate))
	NewTestNotification(t, n, u.ID, TNWithUser(u2.Data.(*models.User)))

	// just to be able to take their addresses
	create := models.UserActionCreate
	upload := models.UserActionUpload
	pTrue := true

	cases := []struct {
		name     string
		recp     uuid.UUID
		offset   int
		limit    int
		action   *models.UserAction
		userID   *uuid.UUID
		targetID *uuid.UUID
		seen     *bool
		count    int
		filter   int
	}{
		{
			name:   "all",
			count:  10,
			filter: 10,
		},
		{
			name:   "first 5",
			offset: 0,
			limit:  5,
			count:  10,
			filter: 5,
		},
		{
			name:   "last 5",
			offset: 5,
			limit:  5,
			count:  10,
			filter: 5,
		},
		{
			name:   "first 5 after [2]",
			offset: 3,
			limit:  5,
			count:  10,
			filter: 5,
		},
		{
			name:   "action upload",
			action: &upload,
			count:  9,
			filter: 9,
		},
		{
			name:   "action create",
			action: &create,
			count:  1,
			filter: 1,
		},
		{
			name:     "with target",
			targetID: &a.ID,
			count:    1,
			filter:   1,
		},
		{
			name:   "with user",
			userID: &u2.ID,
			count:  1,
			filter: 1,
		},
		{
			name:   "with seen",
			seen:   &pTrue,
			count:  0,
			filter: 0,
		},
		{
			name:   "with not seen",
			limit:  5,
			seen:   new(bool),
			count:  10,
			filter: 5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.recp = u.ID

			var ua []models.UserAction
			if tc.action != nil {
				ua = append(ua, *tc.action)
			}
			ns, err := n.Filter(ctx, tc.recp, tc.offset, tc.limit, ua, tc.userID, tc.targetID, tc.seen, nil, nil, nil)
			if err != nil {
				t.Fatalf("Filter failed: %v", err)
			}

			if len(ns) != tc.filter {
				t.Errorf("Expected %d results from filter; got %d: %#v", tc.filter, len(ns), ns)
			}

			count, err := n.Count(ctx, tc.recp, ua, tc.userID, tc.targetID, tc.seen, nil, nil, nil)
			if err != nil {
				t.Fatalf("Count failed: %v", err)
			}
			if count != tc.count {
				t.Errorf("Expected %d results from count; got %d", tc.count, count)
			}
		})
	}
}

func tNotifyList(t *testing.T, notifier Notifier, recp uuid.UUID, a models.UserAction, n int) {
	var action *models.UserAction
	if a != "" {
		action = &a
	}
	ns, err := notifier.List(ctx, recp, action)
	if err != nil {
		t.Fatalf("List(ctx, nil) failed: %v", err)
	}

	if len(ns) != n {
		t.Errorf("Expected %d results; got %d: %#v", n, len(ns), ns)
	}
}

func tNotifySee(t *testing.T, notifier Notifier, n models.Notification) {
	if err := notifier.See(ctx, n.ID, n.RecipientID); err != nil {
		t.Fatalf("failed to mark notification as seen; err: %s", err)
	}

	getn, err := notifier.Get(ctx, n.ID, n.RecipientID)
	if err != nil {
		t.Fatalf("faled to get notification; err: %s", err)
	}

	if getn.Old != n.Old {
		t.Fatalf("expected: %v \n got: %v", n.Old, getn.Old)
	}

	if !getn.Seen {
		t.Fatalf("expected notification to have been marked as seen, got: %v", getn.Seen)
	}
}
