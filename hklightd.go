package main

import (
	"flag"
	"fmt"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"strings"
)

func turnLightOn() {
	log.Println("Turn Light On")
}

func turnLightOff() {
	log.Println("Turn Light Off")
}

func main() {


	info := accessory.Info{
		Name:         	"Lamp",
		Manufacturer: 	"Insteon",
		SerialNumber: 	"2D.83.B4",
		Model:		"On-Off Module",
	}
	lrnw := accessory.NewLightbulb(info)

	doorinfo := accessory.Info{
		Name:         	"Bulb",
		Manufacturer: 	"Insteon",
		Model:		"LED Light Bulb",
	}
	door := accessory.NewLightbulb(doorinfo)

	bedltinfo := accessory.Info{
		Name:         	"Bedroom Light",
		Manufacturer: 	"Insteon",
		Model:		"Fanlinc Dimmer",
	}
	bedlt := accessory.NewLightbulb(bedltinfo)

	closetinfo := accessory.Info{
		Name:         	"Closet Light",
		Manufacturer: 	"Insteon",
		Model:		"LED Light Bulb",
	}
	closet := accessory.NewLightbulb(closetinfo)


	broker := flag.String("broker", "tcp://localhost:1883", "The broker URI. ex: tcp://10.10.1.1:1883")
	password := flag.String("password", "", "The password (optional)")
	user := flag.String("user", "", "The User (optional)")
	id := flag.String("id", "testgoid", "The ClientID (optional)")
	cleansess := flag.Bool("clean", false, "Set Clean Session (default false)")
	qos := flag.Int("qos", 0, "The Quality of Service 0,1,2 (default 0)")
	store := flag.String("store", ":memory:", "The Store Directory (default use memory store)")
	flag.Parse()


	opts := MQTT.NewClientOptions()
	opts.AddBroker(*broker)
	opts.SetClientID(*id)
	opts.SetUsername(*user)
	opts.SetPassword(*password)
	opts.SetCleanSession(*cleansess)
	if *store != ":memory:" {
		opts.SetStore(MQTT.NewFileStore(*store))
	}


	mqttClient := MQTT.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topic	:= "oh-out/state/LR_NW_Light"
        if token := mqttClient.Subscribe(topic, byte(*qos), func(cl MQTT.Client, msg MQTT.Message){
			pl := string(msg.Payload())
			fmt.Printf("got msg: %s on %s\n", pl, msg.Topic())
			lrnw.Lightbulb.On.SetValue(strings.EqualFold(pl, "ON"))
		}); token.Wait() && token.Error() != nil {
                fmt.Println(token.Error())
                os.Exit(1)
        }

	lrnw.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		msg := "OFF"
		if on == true {
			msg = "ON"
		} 
		log.Printf("Turn Light %s", msg)
		topic := "oh-in/cmd/LR_NW_Light"
		token := mqttClient.Publish(topic, byte(*qos), false, msg)
		token.Wait()
	})

	door.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		msg := "OFF"
		if on == true {
			msg = "ON"
		} 
		log.Printf("Turn Door Light %s", msg)
		topic := "oh-in/cmd/Porch_Led_Bulb"
		token := mqttClient.Publish(topic, byte(*qos), false, msg)
		token.Wait()
	})

	bedlt.Lightbulb.Brightness.OnValueRemoteUpdate(func(b int) {
		log.Printf("Turn Bed Light %d", b)
		topic := "oh-in/cmd/fanLincDimmer"
		token := mqttClient.Publish(topic, byte(*qos), false, fmt.Sprintf("%d", b))
		token.Wait()
	})
	bedlt.Lightbulb.On.OnValueRemoteUpdate(func(b bool) {
		msg := "OFF"
		if b {
			msg = "ON"
		}
		log.Printf("Turn Bed Light %s", msg)
		topic := "oh-in/cmd/fanLincDimmer"
		token := mqttClient.Publish(topic, byte(*qos), false, msg)
		token.Wait()
	})

	closet.Lightbulb.Brightness.OnValueRemoteUpdate(func(b int) {
		log.Printf("Turn Closet Light %d", b)
		topic := "oh-in/cmd/Chad_Closet_Led_Bulb"
		token := mqttClient.Publish(topic, byte(*qos), false, fmt.Sprintf("%d", b))
		token.Wait()
	})

	closet.Lightbulb.On.OnValueRemoteUpdate(func(b bool) {
		msg := "OFF"
		if b {
			msg = "ON"
		}
		log.Printf("Turn Closet Light %s", msg)
		topic := "oh-in/cmd/Chad_Closet_Led_Bulb"
		token := mqttClient.Publish(topic, byte(*qos), false, msg)
		token.Wait()
	})

	port := 31330
	lrnw_config := hc.Config{Port: fmt.Sprintf("%d", port), Pin: "12344321", StoragePath: "/var/lib/hc/db" }

	t1, err := hc.NewIPTransport(lrnw_config, lrnw.Accessory, door.Accessory, bedlt.Accessory, closet.Accessory)
	if err != nil {
		log.Fatal(err)
	}
	port++

/*
	door_config := hc.Config{Port: fmt.Sprintf("%d", port), Pin: "12344321", StoragePath: "/var/lib/hc/door" }
	t2, err := hc.NewIPTransport(door_config, door.Accessory)
	if err != nil {
		log.Fatal(err)
	}
	port++
*/

	hc.OnTermination(func() {
		t1.Stop()
		//t2.Stop()
		mqttClient.Disconnect(250)
		fmt.Println("Disconnecting from mqtt broker")
	})


	t1.Start()
	fmt.Println("Started transport")
	//t2.Start()
	//fmt.Println("Started transport")
}
