package drivers

import (
	"errors"
	"math"
	"time"
)

var temperatureCorrection = []float64{2.2619e-5, -0.01479, 1.56}
var humidityCorrection = -2.24e-1

var gasParas = map[string][]float64{
	"CO2": {5.30, -0.34, 414},
	"Tol": {3.49, -0.32, 0.0487},
}

const calibaraionSampleTimes = 100
const calibrationSampleInterval = 500
const readSampleInterval = 50
const readSampleTimes = 5

type MQ135 struct {
	Mcp         *MCP3008 // MCP3008
	Vcc         float64  // for mq135 voltage
	Gas         string   // for mq135: CO2、NH3、CO、EtOH、Tol、Ace
	Ro          float64  // for mq135
	Rl          float64  // for mq135
	Temperature float64  // for mq135
	Humidity    float64  // for mq135
}

func NewMQ135(mcp *MCP3008, vcc float64, gas string, ro float64, rl float64) *MQ135 {
	mq135 := &MQ135{Mcp: mcp, Vcc: vcc, Gas: gas, Ro: ro, Rl: rl}
	return mq135
}

func (mq *MQ135) CalibrationRo(temperature float64, humidity float64) float64 {
	val := 0.0
	for i := 0; i < calibaraionSampleTimes; i++ {
		val += mq.MeasureResistance(temperature, humidity)
		time.Sleep(time.Millisecond * calibrationSampleInterval)
	}

	val = val / float64(calibaraionSampleTimes)

	//fmt.Println("Rs:", val)

	//R0 = RS * (1 / A * c)-1/B
	mq.Ro = val *
		math.Pow(gasParas[mq.Gas][2]/
			math.Pow(1.0/gasParas[mq.Gas][0], 1.0/gasParas[mq.Gas][1]),
			-1.0/(1.0/gasParas[mq.Gas][1]))

	return mq.Ro
}

func (mq *MQ135) MeasureGasConcentration(temperature float64, humidity float64) (float64, error) {

	rs := 0.0
	for i := 0; i < readSampleTimes; i++ {
		rs += mq.MeasureResistance(temperature, humidity)
		time.Sleep(time.Millisecond * readSampleInterval)
	}
	rs = rs / float64(readSampleTimes)

	if mq.Ro == 0 {
		return 0, errors.New("ro not set")
	}

	ratio := rs / mq.Ro
	// c =  A * (RS / R0)B
	return math.Pow(1.0/gasParas[mq.Gas][0], 1.0/gasParas[mq.Gas][1]) *
		math.Pow(ratio, 1.0/gasParas[mq.Gas][1]), nil
}

func (mq *MQ135) MeasureResistance(temperature float64, humidity float64) float64 {
	var voltage float64 = mq.MeasureVoltage()
	return (mq.Vcc/voltage - 1) * mq.Rl / mq.correctResistance(temperature, humidity)
}

func (mq *MQ135) correctResistance(temperature float64, humidity float64) float64 {
	a := temperatureCorrection[0]
	b := temperatureCorrection[1]
	c := temperatureCorrection[2]

	return a*math.Pow(temperature, 2) + b*temperature + c + humidityCorrection*(humidity-0.65)
}

func (mq *MQ135) MeasureVoltage() float64 {
	return mq.Mcp.ReadAdc()
}
