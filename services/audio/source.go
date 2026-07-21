package audio

import (
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/common/logger"
	coreaudio "smegg.me/smeggtuner/core/audio"
)

// File playback is paced to the recording's sample rate, not the decode speed.
const filePlaybackRealtime = true

// GTK filter globs are case sensitive, so .WAV needs its own entry.
const dialogFilterGlob = "*.wav;*.WAV"

// SelectMic switches to a capture device ("" = system default); the device is looked up now, so an unplugged one fails here, not at engine start.
func (s *Service) SelectMic(deviceID string) error {
	name, err := micName(deviceID)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.current = SourceDTO{Kind: SourceMic, DeviceID: deviceID, Name: name}
	s.mu.Unlock()
	logger.Debug(logger.MsgAudioSourceSelected,
		logger.Str("kind", string(SourceMic)),
		logger.Str("device_id", deviceID),
		logger.Str("name", name))
	return nil
}

// SelectFile decodes the WAV once here (a bad path fails now, not at engine start) and keeps it as the transport the view drives.
func (s *Service) SelectFile(path string, loop bool) error {
	file, err := coreaudio.NewFileSource(path, filePlaybackRealtime, loop)
	if err != nil {
		logger.Warn(logger.MsgAudioFileUnreadable, logger.Str("path", path), logger.Err(err))
		return ErrFileUnreadable
	}
	s.mu.Lock()
	s.file = file
	file.SetSink(sinkOrNil(s.speakerLocked()))
	s.current = SourceDTO{
		Kind: SourceFile,
		Path: path,
		Loop: loop,
		Name: filepath.Base(path),
	}
	s.mu.Unlock()
	logger.Debug(logger.MsgAudioSourceSelected,
		logger.Str("kind", string(SourceFile)),
		logger.Str("path", path),
		logger.Bool("loop", loop))
	return nil
}

// OpenFileDialog shows the native WAV picker and returns the chosen path, or "" if cancelled or no Wails app (tests). It does not select the file.
func (s *Service) OpenFileDialog(title, filterName string) (string, error) {
	app := application.Get()
	if app == nil {
		logger.Debug(logger.MsgAudioDialogUnavailable)
		return "", nil
	}
	path, err := app.Dialog.OpenFile().
		SetTitle(title).
		CanChooseFiles(true).
		CanChooseDirectories(false).
		AddFilter(filterName, dialogFilterGlob).
		PromptForSingleSelection()
	if err != nil {
		logger.Warn(logger.MsgAudioDialogFailed, logger.Err(err))
		return "", ErrFileUnreadable
	}
	return path, nil
}

// Build returns the Source for the current selection and whether it is a mic (tells services/tuner to calibrate noise); a mic is built fresh, a file is the kept transport.
func (s *Service) Build() (coreaudio.Source, bool, error) {
	current := s.Current()
	if current.Kind == SourceFile {
		// The recording must still exist on disk; a measurement from a vanished file is an uncheckable session row.
		if _, err := os.Stat(current.Path); err != nil {
			logger.Warn(logger.MsgAudioFileUnreadable,
				logger.Str("path", current.Path), logger.Err(err))
			s.mu.Lock()
			s.file = nil
			s.mu.Unlock()
			return nil, false, ErrFileUnreadable
		}

		s.mu.Lock()
		file := s.file
		s.mu.Unlock()
		if file == nil {
			// A path from a config replayed at startup was never decoded.
			var err error
			if file, err = coreaudio.NewFileSource(current.Path, filePlaybackRealtime, current.Loop); err != nil {
				logger.Warn(logger.MsgAudioFileUnreadable,
					logger.Str("path", current.Path), logger.Err(err))
				return nil, false, ErrFileUnreadable
			}
			s.mu.Lock()
			s.file = file
			file.SetSink(sinkOrNil(s.speakerLocked()))
			s.mu.Unlock()
		}
		file.SetLoop(current.Loop)
		return file, false, nil
	}
	// NewMicSource opens nothing; a vanished device surfaces at Start or via Err, both mapped to ErrDeviceGone.
	src, err := coreaudio.NewMicSource(current.DeviceID)
	if err != nil {
		logger.Warn(logger.MsgAudioDeviceGone, logger.Err(err))
		return nil, true, ErrDeviceGone
	}
	return src, true, nil
}
