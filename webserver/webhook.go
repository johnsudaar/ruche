package webserver

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/johnsudaar/ruche/config"
	"github.com/johnsudaar/ruche/influx"
	"github.com/pkg/errors"

	"github.com/Scalingo/go-utils/logger"
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
	ctx := req.Context()
	log := logger.Get(ctx)
	config := config.Get()

	var body Input
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		log.WithError(err).Error("fail to decode body")
		return errors.Wrap(err, "fail to decode body")
	}

	// 00ed0730110000390116000000000000000000

	log.Infof("Decoding %v", body.Value.Payload)
	valueBytes, err := hex.DecodeString(body.Value.Payload)
	if err != nil {
		log.WithError(err).Error("fail to decode payload (hex)")
		return errors.Wrap(err, "fail to decode payload (hex)")
	}

	valueStr := string(valueBytes)

	if valueStr == "Restart" {
		log.Info("Restart: Ignoring...")
		return nil
	}

	values := make(map[string]interface{})
	tags := make(map[string]string)
	values["location_alt"] = body.Location.Alt
	values["location_accuracy"] = body.Location.Accuracy
	values["location_lon"] = body.Location.Lon
	values["location_lat"] = body.Location.Lat

	values = parsePayload(valueBytes)
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

func checkErr(ctx context.Context, err error, value string) error {
	log := logger.Get(ctx)
	log.WithError(err).WithField("field", value).Error(err.Error())
	return err
}

func getValue(payload string, start, end int) string {
	return strings.Trim(payload[start:end], "\x00 ")
}

func getUInt16(value []byte) uint16 {
	res := uint16(value[1])<<8 | uint16(value[0])

	return res
}

func parsePayload(payload []byte) map[string]interface{} {
	values := make(map[string]interface{})

	values["rucher_id"] = payload[0]

	values["temp"] = float64(getUInt16(payload[1:3])) / 100.0
	values["hum"] = float64(getUInt16(payload[3:5])) / 100.0
	values["lum"] = float64(getUInt16(payload[5:7]))
	values["bat_tension"] = float64(getUInt16(payload[7:9])) / 100.0
	values["sol_tension"] = float64(getUInt16(payload[9:11])) / 100.0
	values["mass_r1"] = float64(getUInt16(payload[11:13])) / 100.0
	values["mass_r2"] = float64(getUInt16(payload[13:15])) / 100.0
	values["mass_r3"] = float64(getUInt16(payload[15:17])) / 100.0
	values["mass_r4"] = float64(getUInt16(payload[17:19])) / 100.0

	return values
}
