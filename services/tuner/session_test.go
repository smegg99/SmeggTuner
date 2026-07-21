package tuner

import (
	"errors"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/datastore/datastoretest"
	"smegg.me/smeggtuner/core/dsp"
	coresession "smegg.me/smeggtuner/core/session"
	sessionsvc "smegg.me/smeggtuner/services/session"
)

// sessionService returns a tuner wired to a session service over a throwaway database, on the a-8 fixture.
func sessionService(t *testing.T) (*Service, *sessionsvc.Service) {
	t.Helper()
	datastoretest.Init(t)
	sessions := sessionsvc.New()
	t.Cleanup(func() { _ = sessions.ServiceShutdown() })

	s := fixtureService(t)
	s.sessions = sessions
	return s, sessions
}

func openSession(t *testing.T, sessions *sessionsvc.Service, reeds int, a4 float64) *sessionsvc.SessionDTO {
	t.Helper()
	dto, err := sessions.Create(sessionsvc.NewSessionDTO{
		Name:       "Morino",
		Instrument: coresession.Instrument{ReedCount: reeds, A4: a4},
	})
	if err != nil {
		t.Fatal(err)
	}
	return dto
}

func TestOpenSessionAppliesA4AndReedCount(t *testing.T) {
	s, sessions := sessionService(t)
	// Read the app's own setting, not assume it: common/config is process-global and other tests move it.
	base := s.snapshot()

	openSession(t, sessions, 2, 442)

	cfg := s.snapshot()
	if cfg.A4 != 442 || cfg.ReedCount != 2 {
		t.Fatalf("engine = a4 %v / reeds %d, want the session's 442 / 2", cfg.A4, cfg.ReedCount)
	}
	if st := s.Settings(); st.A4 != 442 || st.ReedCount != 2 || st.SessionReeds != 2 {
		t.Fatalf("settings = %+v, want the session's", st)
	}

	// Closing gives the app its own reference back: the next instrument must not inherit this one's.
	if err := sessions.Close(); err != nil {
		t.Fatal(err)
	}
	if cfg := s.snapshot(); cfg.A4 != base.A4 || cfg.ReedCount != base.ReedCount {
		t.Fatalf("engine = a4 %v / reeds %d with nothing open, want the app's own %v / %d",
			cfg.A4, cfg.ReedCount, base.A4, base.ReedCount)
	}
	if st := s.Settings(); st.SessionReeds != 0 {
		t.Fatalf("settings = %+v, want no session reeds with nothing open", st)
	}
}

func TestFiveReedSessionTracksAllFive(t *testing.T) {
	s, sessions := sessionService(t)

	dto := openSession(t, sessions, 5, 440)
	if dto.Instrument.ReedCount != 5 {
		t.Fatalf("instrument reeds = %d, want the 5 it declares", dto.Instrument.ReedCount)
	}

	if cfg := s.snapshot(); cfg.ReedCount != 5 {
		t.Fatalf("engine reeds = %d, want the register's 5", cfg.ReedCount)
	}
	st := s.Settings()
	if st.ReedCount != 5 || st.SessionReeds != 5 {
		t.Fatalf("settings = %+v, want the engine and the instrument both at 5", st)
	}
}

// With a session open the A4 control routes into the session, which persists it and refuses it once a pass is recorded.
func TestSetA4RoutesIntoTheActiveSession(t *testing.T) {
	s, sessions := sessionService(t)
	openSession(t, sessions, 3, 440)

	if err := s.SetA4(442); err != nil {
		t.Fatal(err)
	}
	if got := sessions.Goal().A4; got != 442 {
		t.Fatalf("session A4 = %v, want the tuner's control to have set it to 442", got)
	}
	if cfg := s.snapshot(); cfg.A4 != 442 {
		t.Fatalf("engine A4 = %v, want 442", cfg.A4)
	}

	// A reading pins the reference: from here the A4 may not move.
	if err := sessions.UpsertTake(coresession.TakeFrom(dsp.Measurement{Note: 69, ScalePitch: 440}, time.Now())); err != nil {
		t.Fatal(err)
	}
	err := s.SetA4(445)
	if !errors.Is(err, sessionsvc.ErrHasReadings) {
		t.Fatalf("A4 with readings recorded: err = %v, want %s", err, sessionsvc.ErrHasReadings.Key)
	}
	if got := sessions.Goal().A4; got != 442 {
		t.Fatalf("session A4 = %v, want the refused change not to have landed", got)
	}
	// Out of range is still the tuner's own rejection, session or no session.
	if err := s.SetA4(400); !errors.Is(err, ErrInvalidA4) {
		t.Fatalf("A4 400: err = %v, want %s", err, ErrInvalidA4.Key)
	}
}

// With no session the tuner keeps and persists its own reference pitch.
func TestSetA4WithNoSessionStaysWithTheTuner(t *testing.T) {
	s, _ := sessionService(t)

	if err := s.SetA4(443); err != nil {
		t.Fatal(err)
	}
	if cfg := s.snapshot(); cfg.A4 != 443 {
		t.Fatalf("engine A4 = %v, want 443", cfg.A4)
	}
	if st := s.Settings(); st.A4 != 443 || st.SessionReeds != 0 {
		t.Fatalf("settings = %+v, want the app's own A4 and no session", st)
	}
}

// A session opened while the engine runs reaches it on the stream, within a heartbeat.
func TestRunningEngineAdoptsASessionOpenedUnderIt(t *testing.T) {
	s, sessions := sessionService(t)
	log := captureEvents(t, s)

	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	if !waitFor(6*time.Second, func() bool { return len(log.all()) > 2 }) {
		t.Fatal("the engine never streamed")
	}

	openSession(t, sessions, 5, 442)

	ok := waitFor(6*time.Second, func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.imposed.a4 == 442 && s.imposed.reeds == 5
	})
	if !ok {
		t.Fatal("a running engine must adopt the session that was opened under it")
	}
	settings := false
	for _, e := range log.all() {
		if e.name == EventSettings {
			settings = true
		}
	}
	if !settings {
		t.Fatal("the UI has to be told the reference pitch moved under it")
	}
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
}
