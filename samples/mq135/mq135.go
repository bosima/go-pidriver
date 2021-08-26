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

	fmt.Println("Calibrating...")
	ro := 0.0
	sampleTimes := 60
	for i := 0; i < sampleTimes; i++ {
		// use realtime temperature and humidity
		for {
			rh, tmp, err = readTR(dht11)
			if err == nil {
				break
			}
			time.Sleep(time.Second * 2)
		}

		// Gas concentration in natural environment:
		// CO2 414ppm
		// Tol 0.0292ppm=0.11mg/m3、0.039ppm=0.15mg/m3、0.0532ppm=0.20mg/m3
		ro += mq135.MeasureRo(tmp, rh, 414.0)
		time.Sleep(time.Second * 2)
	}
	mq135.Ro = ro / float64(sampleTimes)
	fmt.Printf("ro: %f \n", mq135.Ro)

	// continuou measure gas concentration
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
