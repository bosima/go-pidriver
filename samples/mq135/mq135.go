package main

import (
	"fmt"
	"go-pidriver/drivers"
	"os"
	"strconv"
	"time"
)

func readTR(dht11 *drivers.DHT11) (float64, float64, error) {
	time.Sleep(time.Second * 2)
	rh, tmp, err := dht11.ReadData()

	if err != nil {
		return -1, -1, err
	}

	rh = rh / 100.0
	return rh, tmp, nil
}

func main() {
	drivers.OpenRPi()
	defer func() {
		drivers.CloseRPi()
	}()

	fmt.Println("Collect temperature and humidity...")
	dht11, err := drivers.NewDHT11(4)
	if err != nil {
		fmt.Println(err)
		return
	}

	rh := 0.0
	tmp := 0.0
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 2)
		rh, tmp, err = readTR(dht11)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	fmt.Println("RH:", rh)
	tmpStr := strconv.FormatFloat(tmp, 'f', 1, 64)
	fmt.Println("TMP:", tmpStr)

	mcp, err := drivers.NewMCP3008(0, 0, 5.0)
	if err != nil {
		fmt.Println("----------------------------------")
		fmt.Println(err)
		return
	}
	defer func() {
		mcp.End()
	}()

	mq135 := drivers.NewMQ135(mcp, 5.0, "CO2", 10, 2.0)

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

		fmt.Printf("\rtmp: %f, rh: %f, concentration: %f", tmp, rh, concentration)
		os.Stdout.Sync()

		time.Sleep(time.Millisecond * 2000)
	}
}
