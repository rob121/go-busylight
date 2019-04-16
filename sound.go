package led

type ledSound struct {
	soundID uint8
	volume uint8
}

func NewLedSound() *ledSound {
	return &ledSound{}
}

func (s *ledSound) SetSound(soundID, volume uint8) {
	s.soundID = soundID
	s.volume = volume
}
