package samples

import (
	"fmt"
	"go-pidriver/dht11"
)

func DoDHT11() {
	dht11 := &dht11.DHT11{PinNo: 4}
	err := dht11.Init()
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < 10; i++ {
		rh, tmp, err := dht11.ReadData()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("RH:", rh)
		fmt.Println("TMP:", tmp)
	}

	err = dht11.Close()
	if err != nil {
		fmt.Println(err)
	}
}
