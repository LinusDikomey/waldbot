package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	chart "github.com/wcharczuk/go-chart/v2"
	drawing "github.com/wcharczuk/go-chart/v2/drawing"
)

func dayTimeChart(title string, xValues []float64, yValues []float64, maxY float64) []byte {
	graph := chart.Chart {
		Width: 1280,
		Height: 720,
		Background: chart.Style{
			Padding: chart.Box {
				Top: 80,
				Left: 10,
				Right: 10,
				Bottom: 10,
			},
		},
		ColorPalette: waldColorPalette,
		Title: title,
		YAxis: chart.YAxis {
			Name: "Prozente pro 5 Minuten",
			Range: &chart.ContinuousRange{
				Min: 0.0,
				Max: maxY,
			},
			ValueFormatter: func(v interface{}) string {
				if typed, isTyped := v.(float64); isTyped {
					return chart.FloatValueFormatterWithFormat(typed*100.0, "%0.1f%%")
				}
				return ""
			},
		},
		XAxis: chart.XAxis {
			Name: "Uhrzeit",
			ValueFormatter: func(v interface{}) string {
				if typed, isTyped := v.(float64); isTyped {
					return time.Unix(0, int64(typed)).Format("15:04")
				}
				return "error"
			},
		},
		Series: []chart.Series {
			chart.ContinuousSeries {
				XValues: xValues,
				YValues: yValues,
			},
			chart.AnnotationSeries {
				Name: "Cursed",
			},
		},
	}
	// fill ticks (x axis labels)
	for i := 0; i <= 24; i++ {
		graph.XAxis.Ticks = append(graph.XAxis.Ticks, chart.Tick{Value: float64(i * 60), Label: fmt.Sprintf("%v:00", i)})	
	}

	/*for i, xVal := range(xValues) {
		if i == len(xValues) - 1 { break }
		if typed, isTyped := graph.Series[1].(chart.AnnotationSeries); isTyped {
			fmt.Println("Wtf: #", i)
			typed.Annotations = append(typed.Annotations, chart.Value2 {XValue: xVal, YValue: yValues[i], Label: chart.FloatValueFormatterWithFormat( yValues[i]*100.0, "%0.1f%%")})
			graph.Series[1] = typed
		}
	}*/

	// render
	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		log.Fatal("Error while creating diagram: ", err)
	}
	return buffer.Bytes()
}


var waldColorPalette WaldColorPalette
type WaldColorPalette struct{}

func (dp WaldColorPalette) BackgroundColor() drawing.Color {
	return drawing.Color{R: 54, G: 57, B: 63, A: 255}
	//return drawing.Color(chart.DefaultBackgroundColor)
}

func (dp WaldColorPalette) BackgroundStrokeColor() drawing.Color {
	return drawing.Color(chart.DefaultBackgroundStrokeColor)
}

func (dp WaldColorPalette) CanvasColor() drawing.Color {
	return drawing.Color{R: 54, G: 57, B: 63, A: 255}
	//return drawing.Color(chart.DefaultCanvasColor)
}

func (dp WaldColorPalette) CanvasStrokeColor() drawing.Color {
	return chart.DefaultCanvasStrokeColor
}

func (dp WaldColorPalette) AxisStrokeColor() drawing.Color {
	return drawing.Color{R: 0, G: 0, B: 0, A: 255}
	//return chart.DefaultAxisColor
}

func (dp WaldColorPalette) TextColor() drawing.Color {
	return drawing.Color{R: 255, G: 255, B: 255, A: 255}
	//return chart.DefaultTextColor
}

func (dp WaldColorPalette) GetSeriesColor(index int) drawing.Color {
	return chart.GetDefaultColor(index)
}