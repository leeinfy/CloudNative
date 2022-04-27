package main

import (
	"CloudNative/FinalProject/stockapi"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"google.golang.org/grpc"
)

const (
	serverPort   = "localhost:50001"
	MLenginePort = ":50002"
)

//var MLengineClient stockapi.StockPerdictionClient

func main() {

	//set up http server mux
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(info))
	mux.Handle("/stock", http.HandlerFunc(request))
	//set port to localhost:50001
	log.Print("Setup Server......")
	log.Fatal(http.ListenAndServe(serverPort, mux))
}

// generate data for line chart
func generateLineItems(v []float32) []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < 50; i++ {
		items = append(items, opts.LineData{Value: v[i]})
	}
	return items
}

func info(w http.ResponseWriter, req *http.Request) {

}

func request(w http.ResponseWriter, req *http.Request) {
	// get name from url
	stockName := req.URL.Query().Get("name")
	log.Println("receive request from client ", stockName)
	// Set up a connection to the server.
	log.Println("make connection to ML Engine...")
	conn, err := grpc.Dial(MLenginePort, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Fail to connect to ML Engine: ", err)
	}
	defer conn.Close()
	c := stockapi.NewStockPredictionClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// get current time
	currentTime := time.Now()
	date := fmt.Sprintf("%v-%v-%v", currentTime.Year(), int(currentTime.Month()), currentTime.Day())
	log.Println("Send request from ML Engine .....")
	r, err := c.GetStock(ctx, &stockapi.APIRequest{Name: stockName, Date: date})
	if err != nil {
		log.Println(err)
		return
	}
	if r.Status != "" {
		fmt.Fprintln(w, r.Status)
		return
	}
	log.Println("Plot the chart and show the result....")
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
		AddSeries(stockName, generateLineItems(r.Data)).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: false}))
	line.Render(w)
	fmt.Fprintf(w, "Prediction of stock value: $%v, ", r.Prediction)
	fmt.Fprintf(w, "Our Machine Learning Engine recommandation: %s", r.Recomandation)
	log.Println("request finished....")
}
