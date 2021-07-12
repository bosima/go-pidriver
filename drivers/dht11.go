package drivers

import (
	"errors"
	"go-pidriver/util"
	"strconv"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

type DHT11 struct {
	PinNo        uint8
	pin          *rpio.Pin
	val          []uint8
	closeFunc    func() error
	lastReadTime time.Time
}

func NewDHT11() *DHT11 {
	return &DHT11{PinNo: 4}
}

func (d *DHT11) Init() error {
	err := rpio.Open()
	if err != nil {
		return errors.New("init error: " + err.Error())
	}

	d.closeFunc = func() error {
		return rpio.Close()
	}

	pin := rpio.Pin(d.PinNo)
	d.pin = &pin

	return nil
}

func (d *DHT11) ReadData() (rh float64, tmp float64, err error) {
	if !d.lastReadTime.IsZero() && time.Since(d.lastReadTime).Milliseconds() <= 1000 {
		return -1, -1, errors.New("read interval must be greater than 1 seconds")
	}

	p := d.pin
	d.val = []uint8{0, 0, 0, 0, 0}

	resetDht(p)

	status := checkDhtStatus(p)
	if !status {
		return -1, -1, errors.New("device is not ready")
	}

	// dht output data: 40bit
	for i := 0; i < 5; i++ {
		for k := 0; k < 8; k++ {
			v := readBit(p)
			if v == rpio.High {
				leftLength := 8 - k - 1
				if leftLength > 0 {
					d.val[i] = d.val[i] | 1<<leftLength
				} else {
					d.val[i] = d.val[i] | 1
				}
			}
		}
	}

	if !checkData(d.val) {
		return -1, -1, errors.New("data verification failed")
	}

	rh, err = strconv.ParseFloat(strconv.Itoa(int(d.val[0]))+"."+strconv.Itoa(int(d.val[1])), 32)
	tmp, err = strconv.ParseFloat(strconv.Itoa(int(d.val[2]))+"."+strconv.Itoa(int(d.val[3])), 32)

	d.lastReadTime = time.Now()

	return rh, tmp, err
}

func (d *DHT11) Close() error {
	return d.closeFunc()
}

func resetDht(p *rpio.Pin) {
	p.Output()
	p.High()
	util.Delay(2)

	// send start signal, must great than 18ms
	p.Low()
	util.Delay(25)

	// then over the signal
	p.High()

	// ready to read data
	p.Input()
	p.PullUp()

	// wait 20-40us
	util.DelayMicroseconds(30)
}

func checkDhtStatus(p *rpio.Pin) bool {
	// dht response start: first 80us low, then 80us high
	wait := 0
	for wait < 100 {
		if v := p.Read(); v == rpio.Low {
			break
		}
		util.DelayMicroseconds(1)
		wait += 1
	}
	if wait >= 100 {
		return false
	}

	wait = 0
	for wait < 100 {
		if v := p.Read(); v == rpio.High {
			break
		}
		util.DelayMicroseconds(1)
		wait += 1
	}
	if wait >= 100 {
		return false
	}

	return true
}

func readBit(p *rpio.Pin) rpio.State {
	// for per bit: first 50us low, then 26-70us high
	// 26-28us high represents 0
	// 70us high represents 1
	wait := 0
	for wait < 100 {
		if v := p.Read(); v == rpio.Low {
			break
		}
		util.DelayMicroseconds(1)
		wait += 1
	}

	wait = 0
	for wait < 100 {
		if v := p.Read(); v == rpio.High {
			break
		}
		util.DelayMicroseconds(1)
		wait += 1
	}

	util.DelayMicroseconds(40)

	return p.Read()
}

func checkData(data []uint8) bool {
	sum := 0
	for _, v := range data[:4] {
		sum += int(v)
	}

	if sum != int(data[4]) {
		return false
	}
	return true
}
