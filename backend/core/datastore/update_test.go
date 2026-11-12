// backend/core/datastore/update_test.go
package datastore_test

import (
	"testing"

	"github.com/smegg99/s99wails/v2/shell/update"
	"github.com/smegg99/s99wails/v2/shell/windowstate"

	"smegg.me/smeggtuner/backend/core/datastore"
	"smegg.me/smeggtuner/backend/core/datastore/datastoretest"
	"smegg.me/smeggtuner/backend/core/session"
)

func TestUpdateStoreRoundTrip(t *testing.T) {
	datastoretest.Init(t)
	store := datastore.UpdateStore{}
	if got, ok, err := store.Load(); err != nil || ok || got != (update.Memory{}) {
		t.Fatalf("empty Load = %+v, %v, %v", got, ok, err)
	}
	for _, want := range []update.Memory{
		{Ignored: "1.5.0", Notes: "1.4.3"},
		{Notes: "1.6.0"},
	} {
		if err := store.Save(want); err != nil {
			t.Fatal(err)
		}
		path := datastore.Path()
		if err := datastore.Close(); err != nil {
			t.Fatal(err)
		}
		if err := datastore.Initialize(path); err != nil {
			t.Fatal(err)
		}
		if got, ok, err := store.Load(); err != nil || !ok || got != want {
			t.Fatalf("reopened Load = %+v, %v, %v; want %+v", got, ok, err, want)
		}
	}
}

func TestHasHistory(t *testing.T) {
	t.Run("window", func(t *testing.T) {
		datastoretest.Init(t)
		if datastore.HasHistory() {
			t.Fatal("fresh database has history")
		}
		if err := (datastore.UpdateStore{}).Save(update.Memory{Notes: "1.4.3"}); err != nil {
			t.Fatal(err)
		}
		if datastore.HasHistory() {
			t.Fatal("update memory counts as user history")
		}
		if err := (datastore.WindowStore{}).Save(windowstate.State{Width: 1100, Height: 720}); err != nil {
			t.Fatal(err)
		}
		if !datastore.HasHistory() {
			t.Fatal("saved window is missing from history")
		}
	})
	t.Run("session", func(t *testing.T) {
		datastoretest.Init(t)
		sn := session.New("existing", session.Instrument{ReedCount: 1}, 440)
		if err := datastore.Get().Create(sn).Error; err != nil {
			t.Fatal(err)
		}
		if !datastore.HasHistory() {
			t.Fatal("saved session is missing from history")
		}
	})
}

func TestUpdateStoreReportsDatabaseErrors(t *testing.T) {
	datastoretest.Init(t)
	conn, err := datastore.Get().DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	store := datastore.UpdateStore{}
	if _, ok, err := store.Load(); err == nil || ok {
		t.Fatalf("closed Load = %v, %v", ok, err)
	}
	if err := store.Save(update.Memory{Notes: "1.4.3"}); err == nil {
		t.Fatal("closed database accepted Save")
	}
}
