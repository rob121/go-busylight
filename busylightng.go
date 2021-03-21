package led

import (
	"image/color"
	"log"
	"time"

	"github.com/boombuler/hid"
)

// Device type: BusyLight UC Omega
var BusyLightNGUCOmega DeviceType

// helper struct: "turn off device"
var blNGturnOff = ledAnimationFrame{color: color.Black}

func init() {
	BusyLightNGUCOmega = addDriver(usbDriver{
		/*
			Protocol:
			===================================================================
			[1 byte ] 0x00				(always)
			------------ step -------------------------------------------------
			[1 byte ] 0x10				next step id (always OR 0x10)
			[1 byte ] 0x01				repeat interval
			[3 bytes] 0xFF 0x00 0x00	R/G/B each 8 bit unsigned
			[1 byte ] 0x00				"on" timing
			[1 byte ] 0x00				"off" timing
			[1 byte ] 0x80				sound id and volume
										=> 0x80 + sound_id * 8 + sound_volume
			------------ (repeat) ---------------------------------------------
			=> fill up with 0x00 until len(steps) >= 56
			------------ footer -----------------------------------------------
			[6 bytes] 0x06 0x04 0x55 0xff 0xff 0xff   (always)
			[2 bytes] 0x.. 0x..         checksum: 16 bit unsigned int
										=> sum of all bytes
			===================================================================
		*/
		Name:      "BusyLight UC Omega",
		Type:      &BusyLightNGUCOmega,
		VendorId:  0x27BB,
		ProductId: 0x3BCD,
		Open: func(d hid.Device) (Device, error) {
			return newBusyLightNG(d, func(ani *ledAnimationFrame) error {
				steps := []byte{}
				if ani != nil {
					frame := ani.FirstFrame()
					currentID := uint8(0)

					for {
						// serialize frame
						if frame.GetID() == currentID {
							currentID++
						} else {
							// this seems to be a loop -> abort
							break
						}

						// TODO: validity checks

						nextFrameID := frame.GetID()
						if frame.nextFrame != nil {
							nextFrameID = frame.nextFrame.GetID()
						}
						nextStepByte := byte(nextFrameID) | 0x10

						r, g, b, _ := frame.color.RGBA()

						soundByte := byte(0x80)
						if frame.sound != nil {
							soundByte += frame.sound.soundID * 8
							soundByte += frame.sound.volume
						}

						steps = append(steps, nextStepByte, frame.repeatInterval, byte(r>>8), byte(g>>8), byte(b>>8), frame.onTiming, frame.offTiming, soundByte)

						// select next frame
						if frame.nextFrame == nil {
							break
						} else {
							frame = frame.nextFrame
						}
					}
				} else {
					// just create a keepalive frame
					steps = append(steps, 143, 0, 0, 0, 0, 0, 0, 0)
				}

				for len(steps) < 56 {
					steps = append(steps, 0x00)
				}

				// header and footer
				buffer := []byte{0x00}
				buffer = append(buffer, steps...)
				buffer = append(buffer, 0x06, 0x04, 0x55, 0xff, 0xff, 0xff)

				// calculate checksum
				var checksum uint16
				for _, bufferByte := range buffer {
					checksum += uint16(bufferByte & 0xff)
				}
				buffer = append(buffer, byte(checksum>>8), byte(checksum&0xff))

				// send buffer
				err := d.Write(buffer)

				return err
			}), nil
		},
	})
}

type busylightNGDev struct {
	closeChan chan<- struct{}
	dataChan  chan<- *ledAnimationFrame
	closed    *bool
}

func newBusyLightNG(d hid.Device, updateFn func(ani *ledAnimationFrame) error) *busylightNGDev {
	closeChan := make(chan struct{})
	dataChan := make(chan *ledAnimationFrame)
	ticker := time.NewTicker(9 * time.Second) // If nothing is send after 30 seconds the device turns off.
	closed := false
	go func() {
		var curFrames = &blNGturnOff
		for !closed {
			select {
			case <-ticker.C:
				err := updateFn(nil)
				if err != nil {
					log.Printf("Error updating busylight (keepalive): %s\n", err)
					ticker.Stop()
					closed = true
				}
			case frames := <-dataChan:
				curFrames = frames
				err := updateFn(curFrames)
				if err != nil {
					log.Printf("Error updating busylight: %s\n", err)
					ticker.Stop()
					closed = true
				}
			case <-closeChan:
				ticker.Stop()
				closed = true
				updateFn(&blNGturnOff) // turn off device
			}
		}
		d.Close()
	}()
	return &busylightNGDev{closeChan: closeChan, dataChan: dataChan, closed: &closed}
}

func (d *busylightNGDev) SetKeepActive(v bool) error {
	return ErrKeepActiveNotSupported
}

func (d *busylightNGDev) SetColor(c color.Color) error {
	ani := NewLedAnimation()
	ani.SetColor(c)

	d.dataChan <- ani
	return nil
}

func (d *busylightNGDev) SetAnimation(ani *ledAnimationFrame) error {
	d.dataChan <- ani
	return nil
}

func (d *busylightNGDev) TurnOff() error {
	d.dataChan <- &blNGturnOff
	return nil
}

func (d *busylightNGDev) Close() {
	d.closeChan <- struct{}{}
}

func (d *busylightNGDev) IsClosed() bool {
	return *d.closed
}
