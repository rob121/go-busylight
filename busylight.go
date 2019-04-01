package led

import (
	"image/color"
	"time"

	"github.com/baaazen/go-hid"
)

// Device type: BusyLight UC
var BusyLightUC DeviceType

// Device type: BusyLight Lync
var BusyLightLync DeviceType

// Device type: Kuando BusyLight
var KuandoBusyLight DeviceType

func init() {
	BusyLightUC = addDriver(usbDriver{
		Name:      "BusyLight UC",
		Type:      &BusyLightUC,
		VendorId:  0x27BB,
		ProductId: 0x3BCB,
		Open: func(d hid.Device) (Device, error) {
			return newBusyLight(d, func(c color.Color) {
				r, g, b, _ := c.RGBA()
				d.Write([]byte{0x00, 0x00, 0x00, byte(r >> 8), byte(g >> 8), byte(b >> 8), 0x00, 0x00, 0x00})
			}), nil
		},
	})
	BusyLightLync = addDriver(usbDriver{
		Name:      "BusyLight Lync",
		Type:      &BusyLightLync,
		VendorId:  0x04D8,
		ProductId: 0xF848,
		Open: func(d hid.Device) (Device, error) {
			return newBusyLight(d, func(c color.Color) {
				r, g, b, _ := c.RGBA()
				d.Write([]byte{0x00, 0x00, 0x00, byte(r >> 8), byte(g >> 8), byte(b >> 8), 0x00, 0x00, 0x00})
			}), nil
		},
	})
	KuandoBusyLight = addDriver(usbDriver{
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
										=> 0x80 + sound_id * 0x3 + sound_volume
			------------ (repeat) ---------------------------------------------
			=> fill up with 0x00 until len(steps) >= 56
			------------ footer -----------------------------------------------
			[6 bytes] 0x06 0x04 0x55 0xff 0xff 0xff   (always)
			[2 bytes] 0x.. 0x..         checksum: 16 bit unsigned int
										=> sum of all bytes
			===================================================================
		*/
		Name:      "Kuando BusyLight",
		Type:      &KuandoBusyLight,
		VendorId:  0x27BB,
		ProductId: 0x3BCD,
		Open: func(d hid.Device) (Device, error) {
			return newBusyLight(d, func(c color.Color) {
				r, g, b, _ := c.RGBA()

				steps := []byte{0x10, 0x01, byte(r >> 8), byte(g >> 8), byte(b >> 8), 0x01, 0x00, 0x80}
				for len(steps) < 56 {
					steps = append(steps, 0x00)
				}

				// header and footer
				buffer := []byte{0x00}
				buffer = append(buffer, steps...)
				buffer = append(buffer, 0x06, 0x04, 0x55, 0xff, 0xff, 0xff)

				// calculate checksum
				var checksum uint16 = 0
				for _, bufferByte := range buffer {
					checksum += uint16(bufferByte & 0xff)
				}
				buffer = append(buffer, byte(checksum>>8), byte(checksum&0xff))

				// send buffer
				d.Write(buffer)
			}), nil
		},
	})
}

type busylightDev struct {
	closeChan chan<- struct{}
	colorChan chan<- color.Color
}

func newBusyLight(d hid.Device, setcolorFn func(c color.Color)) *busylightDev {
	closeChan := make(chan struct{})
	colorChan := make(chan color.Color)
	ticker := time.NewTicker(20 * time.Second) // If nothing is send after 30 seconds the device turns off.
	go func() {
		var curColor color.Color = color.Black
		closed := false
		for !closed {
			select {
			case <-ticker.C:
				setcolorFn(curColor)
			case col := <-colorChan:
				curColor = col
				setcolorFn(curColor)
			case <-closeChan:
				ticker.Stop()
				setcolorFn(color.Black) // turn off device
				d.Close()
				closed = true
			}
		}
	}()
	return &busylightDev{closeChan: closeChan, colorChan: colorChan}
}

func (d *busylightDev) SetKeepActive(v bool) error {
	return ErrKeepActiveNotSupported
}

func (d *busylightDev) SetColor(c color.Color) error {
	d.colorChan <- c
	return nil
}
func (d *busylightDev) Close() {
	d.closeChan <- struct{}{}
}
