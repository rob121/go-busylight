package main

import (
  //  "fmt"
    "github.com/rob121/go-busylight"
    "image/color"
    "time"
)

var RED color.RGBA = color.RGBA{0xFF, 0x00, 0x00, 0xFF}
var GREEN color.RGBA = color.RGBA{0,128,0,0xFF}
var OFF color.RGBA = color.RGBA{0,0,0,0}
var light led.DeviceInfo

func main() {
    for devInfo := range led.Devices() {

        light = devInfo
      /*
        dev, err := devInfo.Open()

        fmt.Printf("%#v\n",devInfo)
        fmt.Printf("%#v\n",dev)

        p := devInfo.GetPath()
  
       fmt.Println(p)

        if err != nil {
            fmt.Println(err)
            continue
        }
        defer dev.Close()
        dev.SetColor(GREEN)

        time.Sleep(2 * time.Second) // Wait 2 seconds because the device will turn off once it is closed!
        */

    }

    act,_ := light.Open()

    act.SetColor(GREEN)

    time.Sleep(2 * time.Second)

    act.SetColor(RED)

    time.Sleep(2 * time.Second)

    act.SetColor(OFF)

    time.Sleep(2 * time.Second)

    act.SetColor(GREEN)

    time.Sleep(2 * time.Second)

    act.SetColor(OFF)

    act.Close()




}
