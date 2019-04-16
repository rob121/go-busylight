package led

import (
	"image/color"
)

type ledAnimationFrame struct {
	prevFrame *ledAnimationFrame
	nextFrame *ledAnimationFrame
	color color.Color
	onTiming uint8
	offTiming uint8
	repeatInterval uint8
	sound *ledSound
}

func NewLedAnimation() *ledAnimationFrame {
	return &ledAnimationFrame{}
}

func (f *ledAnimationFrame) FirstFrame() *ledAnimationFrame {
	if f.prevFrame == nil {
		return f
	}
	return f.prevFrame.FirstFrame()
}

func (f *ledAnimationFrame) PrevFrame() *ledAnimationFrame {
	return f.prevFrame
}

func (f *ledAnimationFrame) NextFrame() *ledAnimationFrame {
	return f.nextFrame
}

func (f *ledAnimationFrame) NewFrame() *ledAnimationFrame {
	f.nextFrame = &ledAnimationFrame{prevFrame: f}
	return f.nextFrame
}

func (f *ledAnimationFrame) GetID() uint8 {
	if f.prevFrame == nil {
		return 0
	} 
	return f.prevFrame.GetID() + 1
}

func (f *ledAnimationFrame) SetNextFrame(nextFrame *ledAnimationFrame) {
	f.nextFrame = nextFrame
}

func (f *ledAnimationFrame) SetColor(c color.Color) {
	f.color = c
}

func (f *ledAnimationFrame) SetTiming(onTiming, offTiming uint8) {
	f.onTiming = onTiming
	f.offTiming = offTiming
}

func (f *ledAnimationFrame) SetRepeatInterval(repeatInterval uint8) {
	f.repeatInterval = repeatInterval
}

func (f *ledAnimationFrame) SetSound(sound *ledSound) {
	f.sound = sound
}
