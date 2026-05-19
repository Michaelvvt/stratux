package sensors

import (
	"time"

	"github.com/kidoman/embd"
	"github.com/stratux/stratux/sensors/bmp388"
)

type BMP388 struct {
	sensor      *bmp388.BMP388
	temperature float64
	pressure    float64
	running     bool
}

func NewBMP388(i2cbus *embd.I2CBus) (*BMP388, error) {

	// Probe both standard I2C addresses (set by SDO pin: low=0x76, high=0x77).
	// The .Connected() check reads the CHIP_ID register and validates it's
	// either BMP-388 (0x50) or BMP-390 (0x60), so it doubles as a chip-presence
	// probe at each candidate address.
	var workingAddr byte = 0
	for _, candidate := range []byte{bmp388.Address, bmp388.AddressAlt} {
		probe := bmp388.BMP388{Address: candidate, Bus: i2cbus}
		for n := 0; n < 5; n++ {
			if probe.Connected() {
				workingAddr = candidate
				break
			}
			time.Sleep(time.Millisecond)
		}
		if workingAddr != 0 {
			break
		}
	}
	if workingAddr == 0 {
		return nil, bmp388.ErrNotConnected
	}

	bmp := bmp388.BMP388{Address: workingAddr, Config: bmp388.Config{
		Temperature: bmp388.Sampling8X,
		Pressure:    bmp388.Sampling2X,
		IIR:         bmp388.Coeff0,
	}, Bus: i2cbus}
	err := bmp.Configure(bmp.Config)
	if err != nil {
		return nil, err
	}
	newBmp := BMP388{sensor: &bmp}

	go newBmp.run()
	return &newBmp, nil
}
func (bmp *BMP388) run() {
	bmp.running = true
	clock := time.NewTicker(100 * time.Millisecond)
	for bmp.running {
		for _ = range clock.C {
			var p, _ = bmp.sensor.ReadPressure()
			bmp.pressure = p
			var t, _ = bmp.sensor.ReadTemperature()
			bmp.temperature = t
		}

	}
}

func (bmp *BMP388) Close() {
	bmp.running = false
	bmp.sensor.Config.Mode = bmp388.Sleep
	_ = bmp.sensor.Configure(bmp.sensor.Config)
}

// Temperature returns the current temperature in degrees C measured by the BMP280
func (bmp *BMP388) Temperature() (float64, error) {
	if !bmp.running {
		return 0, bmp388.ErrNotConnected
	}

	return bmp.temperature, nil
}

func (bmp *BMP388) Pressure() (float64, error) {
	if !bmp.running {
		return 0, bmp388.ErrNotConnected
	}
	return bmp.pressure, nil
}
