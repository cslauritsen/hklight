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

	acc := accessory.NewLightbulb(info)

	broker := flag.String("broker", "tcp://hahub.local:1883", "The broker URI. ex: tcp://10.10.1.1:1883")
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

        opts.SetDefaultPublishHandler(func(mqttClient MQTT.Client, msg MQTT.Message) {
		if msg.Topic() == "oh-out/state/LR_NW_Light" {
			pl := string(msg.Payload())
			fmt.Printf("got msg: %s on %s\n", pl, msg.Topic())
			acc.Lightbulb.On.SetValue(strings.EqualFold(pl, "ON"))
		}
        })

	mqttClient := MQTT.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topic	:= "oh-out/state/LR_NW_Light"
        if token := mqttClient.Subscribe(topic, byte(*qos), nil); token.Wait() && token.Error() != nil {
                fmt.Println(token.Error())
                os.Exit(1)
        }

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
