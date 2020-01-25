package main

import (
	"fmt"
	"log"
	"os"
	"time"

	config "./config/local"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// MQTT
var token mqtt.Token

func main() {
	opts := mqtt.NewClientOptions().AddBroker(config.MQTTServer())
	opts.SetClientID(config.DeviceName())
	opts.SetUsername(config.MQTTUser())
	opts.SetPassword(config.MQTTPassword())
	opts.SetDefaultPublishHandler(defaultHandler)

	mqttclient := mqtt.NewClient(opts)
	if token = mqttclient.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		panic(token.Error())
	}

	if token = mqttclient.Subscribe(config.TrackChannel(), 0, defaultHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	bc := widgets.NewBarChart()
	bc.Data = []float64{300, 2000, 5000, 9999}
	bc.Labels = []string{"Red", "Green", "Yellow", "Blue"}
	bc.Title = "Speed"
	bc.SetRect(0, 0, 25, 20)
	bc.BarWidth = 5
	bc.BarColors = []ui.Color{ui.ColorRed, ui.ColorGreen, ui.ColorYellow, ui.ColorBlue}
	bc.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	bc.NumStyles = []ui.Style{ui.NewStyle(ui.ColorBlack)}

	ui.Render(bc)

	table := widgets.NewTable()
	table.Rows = [][]string{
		[]string{"Gopher", "Lap", "Speed", "Position"}, //, "Avg. Speed", "Q1", "Q2"},
		[]string{"Red", "BBB", "CCC"},
		[]string{"Gree", "EEE", "FFF"},
		[]string{"Yellow", "HHH", "III"},
		[]string{"Blue", "HHH", "III"},
	}
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = true
	table.BorderStyle = ui.NewStyle(ui.ColorGreen)
	table.SetRect(0, 20, 100, 31)
	table.FillRow = true
	table.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)

	ui.Render(table)

	p := widgets.NewParagraph()
	p.Text = "Open LED Race!"
	p.SetRect(36, 15, 64, 18)

	ui.Render(p)

	var gauges [4]*widgets.Gauge
	for g := 0; g < 4; g++ {
		gauges[g] = widgets.NewGauge()
		gauges[g].SetRect(30, g*3, 70, 3+3*g)
		gauges[g].Percent = 60
		gauges[g].LabelStyle = ui.NewStyle(ui.ColorBlue)
		gauges[g].BorderStyle.Fg = ui.ColorWhite

		switch g {
		case 0:
			gauges[g].Title = "Red Gopher"
			gauges[g].BarColor = ui.ColorRed
			break
		case 1:
			gauges[g].Title = "Green Gopher"
			gauges[g].BarColor = ui.ColorGreen
			break
		case 2:
			gauges[g].Title = "Yellow Gopher"
			gauges[g].BarColor = ui.ColorYellow
			break
		default:
			gauges[g].Title = "Blue Gopher"
			gauges[g].BarColor = ui.ColorBlue
			break
		}
	}

	ui.Render(gauges[0], gauges[1], gauges[2], gauges[3])

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C
	for {
		select {
		case e := <-uiEvents:
			switch e.ID { // event string/identifier
			case "q", "<C-c>": // press 'q' or 'C-c' to quit
				return
			}
		// use Go's built-in tickers for updating and drawing data
		case <-ticker:
			ui.Render(table)
		}
	}
}

var defaultHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	go echo("[" + msg.Topic() + "] " + string(msg.Payload()))
}