package main

import (
	"flag"
	"fmt"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"log"
)

func turnLightOn() {
	log.Println("Turn Light On")
}

func turnLightOff() {
	log.Println("Turn Light Off")
}

func main() {
	broker := flag.String("broker", "tcp://hahub.local:1883", "The broker URI. ex: tcp://10.10.1.1:1883")
	password := flag.String("password", "", "The password (optional)")
	user := flag.String("user", "", "The User (optional)")
	id := flag.String("id", "testgoid", "The ClientID (optional)")
	cleansess := flag.Bool("clean", false, "Set Clean Session (default false)")
	qos := flag.Int("qos", 0, "The Quality of Service 0,1,2 (default 0)")
	num := flag.Int("num", 1, "The number of messages to publish or subscribe (default 1)")
	payload := flag.String("message", "", "The message text to publish (default empty)")
	store := flag.String("store", ":memory:", "The Store Directory (default use memory store)")
	flag.Parse()

	fmt.Printf("Sample Info:\n")
	fmt.Printf("\tbroker:    %s\n", *broker)
	fmt.Printf("\tclientid:  %s\n", *id)
	fmt.Printf("\tuser:      %s\n", *user)
	fmt.Printf("\tpassword:  %s\n", *password)
	fmt.Printf("\tmessage:   %s\n", *payload)
	fmt.Printf("\tqos:       %d\n", *qos)
	fmt.Printf("\tcleansess: %v\n", *cleansess)
	fmt.Printf("\tnum:       %d\n", *num)
	fmt.Printf("\tstore:     %s\n", *store)

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

	info := accessory.Info{
		Name:         	"Lamp",
		Manufacturer: 	"Insteon",
		SerialNumber: 	"2D.83.B4",
		Model:		"On-Off Module",
	}

	acc := accessory.NewLightbulb(info)

	acc.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		if on == true {
			log.Println("Turn Light On")
			fmt.Println("---- doing publish ----")
			topic := "oh-in/cmd/LR_NW_Light"
			token := mqttClient.Publish(topic, byte(*qos), false, "ON")
			token.Wait()
		} else {
			log.Println("Turn Light Off")
			fmt.Println("---- doing publish ----")
			topic := "oh-in/cmd/LR_NW_Light"
			token := mqttClient.Publish(topic, byte(*qos), false, "OFF")
			token.Wait()
		}
	})

	t, err := hc.NewIPTransport(hc.Config{Pin: "32191123"}, acc.Accessory)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		t.Stop()
		mqttClient.Disconnect(250)
		fmt.Println("Disconnecting from mqtt broker")
	})

	t.Start()
}
