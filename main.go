package main

import (
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var table *widgets.Table
var bc *widgets.BarChart
var gauges [4]*widgets.Gauge
var debug *widgets.Paragraph
var conn *net.UDPConn

func main() {
	var err error
	conn, err = net.ListenUDP("udp", &net.UDPAddr{
		Port: 1053,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		panic(err)
	}

	defer conn.Close()
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	bc = widgets.NewBarChart()
	bc.Data = []float64{300, 2000, 5000, 9999}
	bc.Labels = []string{"Red", "Green", "Yellow", "Blue"}
	bc.Title = "Speed"
	bc.SetRect(0, 0, 25, 20)
	bc.BarWidth = 5
	bc.BarColors = []ui.Color{ui.ColorRed, ui.ColorGreen, ui.ColorYellow, ui.ColorBlue}
	bc.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	bc.NumStyles = []ui.Style{ui.NewStyle(ui.ColorBlack)}

	ui.Render(bc)

	table = widgets.NewTable()
	table.Rows = [][]string{
		[]string{"Gopher", "Lap", "Speed", "Distance"}, //, "Avg. Speed", "Q1", "Q2"},
		[]string{"Red", "0", "0", "0"},
		[]string{"Green", "0", "0", "0"},
		[]string{"Yellow", "0", "0", "0"},
		[]string{"Blue", "0", "0", "0"},
	}
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = true
	table.BorderStyle = ui.NewStyle(ui.ColorGreen)
	table.SetRect(0, 20, 70, 31)
	table.FillRow = true
	table.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)

	ui.Render(table)

	p := widgets.NewParagraph()
	p.Text = "Open LED Race!"
	p.SetRect(36, 15, 64, 18)

	ui.Render(p)

	debug = widgets.NewParagraph()
	debug.Text = "DEBUG"
	debug.SetRect(0, 31, 70, 40)

	ui.Render(debug)

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
	go updateFromUDP()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID { // event string/identifier
			case "q", "<C-c>": // press 'q' or 'C-c' to quit
				return
			}
		// use Go's built-in tickers for updating and drawing data
		case <-ticker:
			//updateFromUDP()
		}
	}
}

func updateFromUDP() {
	for {
		message := make([]byte, 1000)
		// don't handle errors #YOLO
		rlen, _, _ := conn.ReadFromUDP(message[:])

		dataStr := strings.TrimSpace(string(message[:rlen]))
		debug.Text = dataStr
		if dataStr != "none" && dataStr != "boot" && dataStr != "AT+C"{
			data := strings.Split(dataStr, "|")
			for i := 0; i < len(data); i++ {
				player := strings.Split(data[i], ",")
				if len(player)<3 {
					continue
				}
				position, _ := strconv.Atoi(player[1])
				position = position / 9
				gauges[i].Percent = position
				speed, _ := strconv.Atoi(player[0])
				if speed < 0 {
					speed = 0
				}
				bc.Data[i] = float64(speed)
				table.Rows[i+1][1] = player[2]
				table.Rows[i+1][2] = strconv.Itoa(speed)
				table.Rows[i+1][3] = strconv.Itoa(position)
			}
		}
		ui.Render(table, bc, gauges[0], gauges[1], gauges[2], gauges[3], debug)
	}

}
