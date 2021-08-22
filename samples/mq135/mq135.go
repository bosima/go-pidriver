package main

import (
	"fmt"
	"go-pidriver/drivers"
	"os"
	"time"
)

func readTR(dht11 *drivers.DHT11) (float64, float64, error) {
	rh, tmp, err := dht11.ReadData()

	if err != nil {
		return 0.0, 0.0, err
	}

	rh = rh / 100.0
	return rh, tmp, nil
}

func main() {
	fmt.Println("-----------raspberry pi------------")
	drivers.OpenRPi()
	defer func() {
		drivers.CloseRPi()
	}()
	fmt.Println("device memory mapping completed")

	fmt.Println("-----------dht11------------")
	dht11, err := drivers.NewDHT11(4)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Collect temperature and humidity...")
	rh := 0.0
	tmp := 0.0
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 2)
		rh, tmp, err = readTR(dht11)
		if err != nil {
			continue
		}
	}
	fmt.Printf("RH: %f \n", rh)
	fmt.Printf("TMP: %f \n", tmp)

	if rh == 0.0 || tmp == 0.0 {
		fmt.Println("can not get temperature or humidity")
		return
	}

	fmt.Println("-----------mcp3008------------")
	mcp, err := drivers.NewMCP3008(0, 0, 5.0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		mcp.End()
	}()
	fmt.Println("mcp3008 started")

	fmt.Println("-----------mq135------------")
	mq135 := drivers.NewMQ135(mcp, 5.0, "CO2", 10, 2.0)

	voltage := mq135.MeasureVoltage()
	fmt.Printf("voltage: %f \n", voltage)

	fmt.Println("Calibrating...")
	ro := mq135.CalibrationRo(tmp, rh)
	fmt.Printf("ro: %f \n", ro)

	for {
		rh, tmp, err := readTR(dht11)
		if err != nil {
			time.Sleep(time.Second * 2)
			continue
		}

		concentration, err := mq135.MeasureGasConcentration(tmp, rh)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("\rtmp: %.2f, rh: %.2f, concentration: %f", tmp, rh, concentration)
		os.Stdout.Sync()

		time.Sleep(time.Millisecond * 2000)
	}
}
