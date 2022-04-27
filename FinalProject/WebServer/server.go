package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

const (
	serverPort   = "localhost:50001"
	MLenginePort = ":50002"
)

//var MLengineClient stockapi.StockPerdictionClient

func main() {
	// Set up a connection to the server.
	/*conn, err := grpc.Dial(MLenginePort, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Fail to connect to MLengine: %v", err)
	}
	defer conn.Close()
	MLengineClient = stockapi.NewStockPerdictionClient(conn)*/
	//set up http server mux
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(info))
	mux.Handle("/stock", http.HandlerFunc(request))
	//set port to localhost:50001
	log.Print("Setup Server......")
	log.Fatal(http.ListenAndServe(serverPort, mux))
}

// generate random data for line chart
func generateLineItems() []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < 50; i++ {
		items = append(items, opts.LineData{Value: rand.Intn(300)})
	}
	return items
}

func info(w http.ResponseWriter, req *http.Request) {

}

func request(w http.ResponseWriter, req *http.Request) {
	// get name from url
	stockName := req.URL.Query().Get("name")
	// get current time
	currentTime := time.Now()
	//date := fmt.Sprintf("%v-%v-%v", currentTime.Year(), int(currentTime.Month()), currentTime.Day())
	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	str := fmt.Sprintf("value of %s in last 50 days, Time: ", stockName) + fmt.Sprintf(currentTime.Format("01-02-2006 15:04:05 Mon"))
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title:    stockName,
			Subtitle: str,
		}))

	// Put data into instance
	var XAxis []string
	for i := -49; i <= 0; i++ {
		month := currentTime.AddDate(0, 0, i).Month()
		day := currentTime.AddDate(0, 0, i).Day()
		XAxis = append(XAxis, fmt.Sprintf("%v-%v", month, day))
	}
	line.SetXAxis(XAxis).
		AddSeries(stockName, generateLineItems()).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	line.Render(w)
	fmt.Fprintf(w, "Our Machine Learning Engine recommandation: \n")
}
