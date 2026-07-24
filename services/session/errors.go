package session

import (
	audiosvc "smegg.me/smeggtuner/services/audio"
)

// ServiceError is the error shape every service hands the frontend: an i18n key the UI translates.
// It is the audio service's type, so errors.As reaches any service's keys through any other.
type ServiceError = audiosvc.ServiceError

var (
	// ErrNoSession reports that the call needs an open session and none is.
	ErrNoSession = &ServiceError{Key: "session.error.noSession"}
	// ErrNotFound reports a session id that is not in the store.
	ErrNotFound = &ServiceError{Key: "session.error.notFound"}
	// ErrLoadFailed reports a session file that would not read.
	ErrLoadFailed = &ServiceError{Key: "session.error.loadFailed"}
	// ErrSaveFailed reports a session that would not write.
	ErrSaveFailed = &ServiceError{Key: "session.error.saveFailed"}
	// ErrNotAnImage reports a file that is not an image.
	ErrNotAnImage = &ServiceError{Key: "session.error.notAnImage"}
	// ErrImageUnreadable reports a photograph that would not open.
	ErrImageUnreadable = &ServiceError{Key: "session.error.imageUnreadable"}
	// ErrInstrumentName reports an instrument saved without a name to find it again by.
	ErrInstrumentName = &ServiceError{Key: "session.error.instrumentName"}
	// ErrInvalidName reports an empty session name.
	ErrInvalidName = &ServiceError{Key: "session.error.invalidName"}
	// ErrInvalidA4 reports a reference pitch outside 430..450 Hz.
	ErrInvalidA4 = &ServiceError{Key: "session.error.invalidA4"}
	// ErrInvalidReedCount reports a reed count outside 1..8.
	ErrInvalidReedCount = &ServiceError{Key: "session.error.invalidReedCount"}
	// ErrInvalidInstrument reports an instrument that failed validation.
	ErrInvalidInstrument = &ServiceError{Key: "session.error.invalidInstrument"}
	// ErrNoRegister reports a register the instrument on the bench does not have.
	ErrNoRegister = &ServiceError{Key: "session.error.noRegister"}
	// ErrNoBassMachine reports a bench turned toward a bass side the instrument never declared.
	ErrNoBassMachine = &ServiceError{Key: "session.error.noBassMachine"}
	// ErrNoBassRegister reports a bass register the instrument does not have.
	ErrNoBassRegister = &ServiceError{Key: "session.error.noBassRegister"}
	// ErrInvalidNote reports a note outside the tuning range.
	ErrInvalidNote = &ServiceError{Key: "session.error.invalidNote"}
	// ErrInvalidReed reports a reed index the curve does not describe.
	ErrInvalidReed = &ServiceError{Key: "session.error.invalidReed"}
	// ErrInvalidUnit reports an authoring unit that is neither cent nor hz.
	ErrInvalidUnit = &ServiceError{Key: "session.error.invalidUnit"}
	// ErrInvalidValue reports a value that is not a pitch at that note.
	ErrInvalidValue = &ServiceError{Key: "session.error.invalidValue"}
	// ErrHasReadings reports a change refused because the session already holds readings: its reference and reed count are then frozen.
	ErrHasReadings = &ServiceError{Key: "session.error.hasReadings"}
	// ErrTakeNotFound reports a reading index the session does not have.
	ErrTakeNotFound = &ServiceError{Key: "session.error.takeNotFound"}
	// ErrNoReadings reports a fit asked of a session with nothing recorded in it.
	ErrNoReadings = &ServiceError{Key: "session.error.noReadings"}
	// ErrFitFailed reports a fit that produced no usable curve.
	ErrFitFailed = &ServiceError{Key: "session.error.fitFailed"}
	// ErrNoCurve reports an import from a session that has no goal curve.
	ErrNoCurve = &ServiceError{Key: "session.error.noCurve"}
)
