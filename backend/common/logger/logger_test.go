// backend/common/logger/logger_test.go
package logger

import (
	"context"
	"testing"

	"github.com/smegg99/s99logger"

	"smegg.me/smeggtuner/backend/common/i18n"
)

type records struct{ values []s99logger.Record }

func (s *records) Write(_ context.Context, record s99logger.Record) error {
	s.values = append(s.values, record)
	return nil
}

func TestLogLanguageChangesWithoutChangingEventIdentity(t *testing.T) {
	defer i18n.SetLocale("en")
	sink := &records{}
	log := s99logger.New(sink, s99logger.Options{Translator: i18n.Translator()})
	event := s99logger.NewEvent(i18n.KeyLogSessionCreated, s99logger.String("session_id", "abc"))
	for _, language := range []string{"en", "pl"} {
		i18n.SetLocale(language)
		log.Info(event)
	}
	if len(sink.values) != 2 {
		t.Fatalf("got %d records", len(sink.values))
	}
	for i, want := range []string{"Session created", "Utworzono sesję"} {
		record := sink.values[i]
		if record.Message != want || record.MessageID != i18n.KeyLogSessionCreated {
			t.Fatalf("record %d: message=%q id=%q", i, record.Message, record.MessageID)
		}
		if len(record.Attrs) != 1 || record.Attrs[0].Key != "session_id" || record.Attrs[0].Value != "abc" {
			t.Fatalf("record %d lost structured context: %+v", i, record.Attrs)
		}
	}
}
