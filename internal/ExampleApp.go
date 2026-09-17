package internal

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/go-beans/go/ioc"
	"github.com/go-external-config/go/str"
	"github.com/go-fyne-timeseries/go/chart"
	"github.com/go-fyne-timeseries/go/chart/colors"
	"github.com/go-jang/go/lang"
	"github.com/go-jang/go/util/objects"
	"github.com/go-jang/go/util/stream"
)

var _ ioc.ApplicationRunner = (*ExampleApp)(nil)

type ExampleApp struct {
	timeSeriesProvider *TimeSeriesProvider `inject:""`
	colorGen           *colors.Generator
}

func NewVisualParamApp() *ExampleApp {
	return &ExampleApp{
		colorGen: colors.NewGenerator(0.45),
	}
}

func (this *ExampleApp) Run(args []string) {
	paramIds := stream.Map(stream.From(args[1:]), func(s string) int64 { return str.Parse[int64](s) }).ToSlice()
	lang.Assert(len(paramIds) > 0, "Wrong arguments count. Example usage: go run ./cmd/app <param>[ <param>]")
	tsc := this.initTimeSeriesColection(paramIds)

	app := app.New()
	win1 := app.NewWindow("Chart")
	win1.SetContent(chart.NewChartPanel(chart.NewChart("Parameters", "", "", tsc, chart.Linear, true, false, false, true, time.Hour)))
	win1.Resize(fyne.NewSize(800, 600))
	win1.CenterOnScreen()
	win1.Show()

	this.timeSeriesProvider.GetHistoricalDataAsync(ioc.Context(), paramIds, time.Now().AddDate(0, 0, -1), time.Now(), tsc)
	this.timeSeriesProvider.Subscribe(ioc.Context(), paramIds, tsc)

	app.Run()
}

func (this *ExampleApp) initTimeSeriesColection(paramIds []int64) *chart.TimeSeriesCollection {
	tsc := chart.NewTimeSeriesCollection()
	for _, paramId := range paramIds {
		ts := chart.TimeSeriesOfName(objects.ToString(paramId), this.colorGen.Next())
		ts.SetMaxItemAge(time.Hour)
		tsc.AddSeries(ts)
	}
	return tsc
}
