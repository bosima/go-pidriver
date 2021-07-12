package main

import (
	"fmt"
	"go-pidriver/drivers"
	"strconv"
	"time"
)

func main() {
	doDHT11()
}

func doDHT11() {
	// 1、init
	dht11 := drivers.NewDHT11()
	err := dht11.Init()
	if err != nil {
		fmt.Println("----------------------------------")
		fmt.Println(err)
		return
	}

	// 2、read
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 2)
		rh, tmp, err := dht11.ReadData()
		if err != nil {
			fmt.Println("----------------------------------")
			fmt.Println(err)
			continue
		}

		fmt.Println("----------------------------------")
		fmt.Println("RH:", rh)

		tmpStr := strconv.FormatFloat(tmp, 'f', 1, 64)
		fmt.Println("TMP:", tmpStr)

	}

	// 3、close
	err = dht11.Close()
	if err != nil {
		fmt.Println(err)
	}
}
