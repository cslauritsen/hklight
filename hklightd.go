package main

import (
	"flag"
	"fmt"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"log"
	"sort"
	"strings"
	"strconv"
)

func main() {

	lightbulbs := make(map[string]*accessory.Lightbulb)

	lightbulbs["LR_NW_Light"] = accessory.NewLightbulb(accessory.Info{
		Name:         "Lamp",
		Manufacturer: "Insteon",
		SerialNumber: "2D.83.B4",
		Model:        "On-Off Module",
	})

	lightbulbs["Porch_Led_Bulb"] = accessory.NewLightbulb(accessory.Info{
		Name:         "Bulb",
		Manufacturer: "Insteon",
		Model:        "LED Light Bulb",
	})

	lightbulbs["fanLincDimmer"] = accessory.NewLightbulb(accessory.Info{
		Name:         "Bedroom Light",
		Manufacturer: "Insteon",
		Model:        "Fanlinc Dimmer",
	})

	lightbulbs["Chad_Closet_Led_Bulb"] = accessory.NewLightbulb(accessory.Info{
			Name:         "Closet Light",
			Manufacturer: "Insteon",
			Model:        "LED Light Bulb",
	})

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

	var keys []string
	for k, _ := range lightbulbs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		k2 := k
		v2 := lightbulbs[k2]
		// subscribe to mqtt topic to get current state of device
		// and keep home kit devices updated when changes are made
		// outside of homekit
		fmt.Printf("Setting up %s\n", k2)
		if token := mqttClient.Subscribe("oh-retain/state/"+k2, byte(*qos), func(cl MQTT.Client, msg MQTT.Message) {
			pl := string(msg.Payload())
			fmt.Printf("got msg: %s on %s\n", pl, msg.Topic())
			if strings.EqualFold(pl, "ON") {
				v2.Lightbulb.On.SetValue(true)
			} else if strings.EqualFold(pl, "OFF") {
				v2.Lightbulb.On.SetValue(false)
			} else {
				// assumption: decimal numeric val?
				bval, err := strconv.Atoi(pl)
				if  err == nil {
					fmt.Printf("Setting brightness %d\n", bval)
					v2.Lightbulb.Brightness.SetValue(bval)
				} else {
					fmt.Printf("did not understand msg: %s: %v\n", pl, err)
				}
			}
		}); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}
		// Handle ON/OFF event from a homekit device
		v2.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
			msg := "OFF"
			if on == true {
				msg = "ON"
			}
			log.Printf("Turn %s %s", k2, msg)
			topic := "oh-in/cmd/" + k2
			token := mqttClient.Publish(topic, byte(*qos), false, msg)
			token.Wait()
		})
		// Handle dimming event from a homekit device
		v2.Lightbulb.Brightness.OnValueRemoteUpdate(func(b int) {
			log.Printf("Turn %s %d", k2, b)
			topic := "oh-in/cmd/" + k2
			token := mqttClient.Publish(topic, byte(*qos), false, fmt.Sprintf("%d", b))
			token.Wait()
		})
	}

	port := 31330
	bridge_config := hc.Config{Port: fmt.Sprintf("%d", port), Pin: "12344321", StoragePath: "/var/lib/hc/db"}

	accessorySlice := make([]*accessory.Accessory, 0)
	for _, k := range keys {
		v := lightbulbs[k]
		a := v.Accessory
		accessorySlice = append(accessorySlice, a)
	}

	var ipt hc.Transport
	var err error
	if len(accessorySlice) > 1 {
		ipt, err = hc.NewIPTransport(bridge_config, accessorySlice[0], accessorySlice[1:]...)
	} else {
		ipt, err = hc.NewIPTransport(bridge_config, accessorySlice[0])
	}
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		ipt.Stop()
		mqttClient.Disconnect(250)
		fmt.Println("Disconnecting from mqtt broker")
	})

	ipt.Start()
	fmt.Println("Started transport")
}

/*
	topic := "oh-out/state/LR_NW_Light"
	if token := mqttClient.Subscribe(topic, byte(*qos), func(cl MQTT.Client, msg MQTT.Message) {
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
*/
