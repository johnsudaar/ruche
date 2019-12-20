package influx

import (
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"gopkg.in/errgo.v1"
)

func Start(influxURL string) (*influx.BatchPoints, error) {
	infos, err := parseConnectionString(influxURL)
	if err != nil {
		return nil, errgo.Mask(err)
	}
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  infos.Database,
		Precision: "s",
	})
	if err != nil {
		return nil, errgo.Mask(err)
	}
	return &bp, nil
}

func Write(influxURL string, bp *influx.BatchPoints) error {
	client, _, err := Client(influxURL)
	if err != nil {
		return errgo.Mask(err)
	}
	defer client.Close()

	err = client.Write(*bp)
	if err != nil {
		return errgo.Mask(err)
	}
	return nil
}

func Add(field string, values map[string]interface{}, tags map[string]string, bp *influx.BatchPoints, time time.Time) error {
	pt, err := influx.NewPoint(field, tags, values, time)
	if err != nil {
		return errgo.Mask(err)
	}
	(*bp).AddPoint(pt)

	return nil
}
