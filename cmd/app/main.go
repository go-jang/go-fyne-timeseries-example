package main

import (
	"github.com/go-beans/go/ioc"
	"github.com/go-fyne-timeseries/example/internal"
)

func init() {
	ioc.Bean[*internal.ExampleApp]().Factory(internal.NewVisualParamApp).Register()
	ioc.Bean[*internal.TimeSeriesProvider]().Factory(internal.NewTimeSeriesProvider).Register()
}

func main() {
	defer ioc.Close()
	ioc.Run()
}
