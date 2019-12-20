package webserver

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Scalingo/go-utils/logger"
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
	Provider string  `json:"provider"`
	Alt      float64 `json:"alt"`
	Accuracy float64 `json:"accuracy"`
	Lon      float64 `json:"lon"`
	Lat      float64 `json:"lat"`
}

type Value struct {
	Payload string `json:"payload"`
}

func Webhook(resp http.ResponseWriter, req *http.Request, params map[string]string) error {
	log := logger.Get(req.Context())
	config := config.Get()

	var body Input
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		log.WithError(err).Error("fail to decode body")
		return errors.Wrap(err, "fail to decode body")
	}

	valueStr, err := hex.DecodeString(body.Value.Payload)
	if err != nil {
		log.WithError(err).Error("fail to decode payload (hex)")
		return errors.Wrap(err, "fail to decode payload (hex)")
	}

	values := make(map[string]interface{})
	tags := make(map[string]string)
	values["location_alt"] = body.Location.Alt
	values["location_accuracy"] = body.Location.Accuracy
	values["location_lon"] = body.Location.Lon
	values["location_lat"] = body.Location.Alt

	if len(valueStr) > 4 {
		log.Info("Model 1 decoding")
		// Model 1: Use only 1 value
		value, err := strconv.ParseFloat(string(valueStr), 32)
		if err != nil {
			log.WithError(err).Error("fail to decode payload (str)")
			return errors.Wrap(err, "fail to decode payload (str)")
		}
		values["temp"] = value
	} else {
		log.Info("Model 2 decoding")

	}

	tags["stream_id"] = body.StreamID
	tags["model"] = body.Model
	tags["location_provider"] = body.Location.Provider
	log.Info(values)
	log.Info(tags)

	bp, err := influx.Start(config.InfluxUrl)
	if err != nil {
		log.WithError(err).Error("fail to open influx connection")
		return errors.Wrap(err, "fail to open influx connection")
	}
	log.Info("Add")

	err = influx.Add("raw", values, tags, bp, body.Created)
	if err != nil {
		log.WithError(err).Error("fail to add batch point")
		return errors.Wrap(err, "fail to add batch point")
	}
	log.Info("Write")

	err = influx.Write(config.InfluxUrl, bp)
	if err != nil {
		log.WithError(err).Error("fail to write points")
		return errors.Wrap(err, "fail to write points")
	}
	log.Info("Done")

	return nil
}
