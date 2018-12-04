package main

import (
	"log"
	"os"

	"github.com/wcharczuk/go-chart"
)

func chartRender(filename string, input *ChartData, output *ChartData) error {

	log.Printf("chartRender: input data points:  %d/%d", len(input.XValues), len(input.YValues))
	log.Printf("chartRender: output data points: %d/%d", len(output.XValues), len(output.YValues))

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
			TickPosition:   chart.TickPositionBetweenTicks,
			ValueFormatter: chart.TimeMinuteValueFormatter,
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
			chart.TimeSeries{
				Name:    "Input",
				XValues: input.XValues,
				YValues: input.YValues,
			},
			chart.TimeSeries{
				Name:    "Output",
				YAxis:   chart.YAxisSecondary,
				XValues: output.XValues,
				YValues: output.YValues,
			},
		},
	}

	return graph.Render(chart.PNG, out)
}
