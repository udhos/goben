package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/util"
)

func chartTime(t time.Time) float64 {
	return util.Time.ToFloat64(t)
}

func chartRender(filename string, input *chartData, output *chartData) error {

	log.Printf("chartRender: input data points:  %d/%d", len(input.xValues), len(input.yValues))
	log.Printf("chartRender: output data points: %d/%d", len(output.xValues), len(output.yValues))

	out, errCreate := os.Create(filename)
	if errCreate != nil {
		return errCreate
	}
	defer out.Close()

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name: "Time",
			Style: chart.Style{
				Show: true, //enables / displays the x-axis
			},
			TickPosition: chart.TickPositionBetweenTicks,
			ValueFormatter: func(v interface{}) string {
				typed := v.(float64)
				typedDate := util.Time.FromFloat64(typed)
				return fmt.Sprintf("%02d:%02d:%02d", typedDate.Hour(), typedDate.Minute(), typedDate.Second())
			},
		},
		YAxis: chart.YAxis{
			Name: "Mbps",
			Style: chart.Style{
				Show: true, //enables / displays the y-axis
			},
		},
		YAxisSecondary: chart.YAxis{
			Style: chart.Style{
				Show: true, //enables / displays the secondary y-axis
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name:    "Input",
				XValues: input.xValues,
				YValues: input.yValues,
			},
			chart.ContinuousSeries{
				Name:    "Output",
				YAxis:   chart.YAxisSecondary,
				XValues: output.xValues,
				YValues: output.yValues,
			},
		},
	}

	graph.Render(chart.PNG, out)

	return nil
}
