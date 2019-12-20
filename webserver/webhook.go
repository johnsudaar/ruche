package webserver

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/johnsudaar/ruche/config"
	"github.com/johnsudaar/ruche/influx"
	"github.com/pkg/errors"
)

type Input struct {
	StreamID string    `json:"streamId"`
	Model    string    `json:"model"`
	Created  time.Time `json:"created"`
	Location Location  `json:"location"`
	Value    Value     `json:"value"`
}

type Location struct {
	Provider string `json:"provider"`
	Alt      int    `json:"alt"`
	Accuracy int    `json:"accuracy"`
	Lon      int    `json:"lon"`
	Lat      int    `json:"lat"`
}

type Value struct {
	Payload string `json:"payload"`
}

func Webhook(resp http.ResponseWriter, req *http.Request, params map[string]string) error {
	config := config.Get()

	var body Input
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		return errors.Wrap(err, "fail to decode body")
	}

	valueStr, err := hex.DecodeString(body.Value.Payload)
	if err != nil {
		return errors.Wrap(err, "fail to decode payload (hex)")
	}

	value, err := strconv.ParseFloat(string(valueStr), 32)
	if err != nil {
		return errors.Wrap(err, "fail to decode payload (hex)")
	}

	values := make(map[string]interface{})
	tags := make(map[string]string)
	values["payload"] = value
	values["location_alt"] = body.Location.Alt
	values["location_accuracy"] = body.Location.Accuracy
	values["location_lon"] = body.Location.Lon
	values["location_lat"] = body.Location.Alt

	tags["stream_id"] = body.StreamID
	tags["model"] = body.Model
	tags["location_provider"] = body.Location.Provider

	bp, err := influx.Start(config.InfluxUrl)
	if err != nil {
		return errors.Wrap(err, "fail to open influx connection")
	}

	err = influx.Add("raw", values, tags, bp, body.Created)
	if err != nil {
		return errors.Wrap(err, "fail to add batch point")
	}

	err = influx.Write(config.InfluxUrl, bp)
	if err != nil {
		return errors.Wrap(err, "fail to write points")
	}

	return nil
}
