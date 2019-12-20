package influx

import (
	"net/url"

	influx "github.com/influxdata/influxdb/client/v2"

	"gopkg.in/errgo.v1"
)

type influxInfo struct {
	Host             string
	User             string
	Password         string
	Database         string
	ConnectionString string
}

func Client(influxURL string) (influx.Client, *influxInfo, error) {
	infos, err := parseConnectionString(influxURL)
	if err != nil {
		return nil, nil, errgo.Mask(err)
	}
	client, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     infos.Host,
		Username: infos.User,
		Password: infos.Password,
	})

	if err != nil {
		return nil, nil, errgo.Mask(err)
	}

	return client, infos, err
}

func parseConnectionString(con string) (*influxInfo, error) {
	url, err := url.Parse(con)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	password, _ := url.User.Password()

	return &influxInfo{
		Host:             url.Scheme + "://" + url.Host,
		User:             url.User.Username(),
		Password:         password,
		Database:         url.Path[1:],
		ConnectionString: con,
	}, nil
}
