package webserver

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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
	ctx := req.Context()
	log := logger.Get(ctx)
	config := config.Get()

	var body Input
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		log.WithError(err).Error("fail to decode body")
		return errors.Wrap(err, "fail to decode body")
	}

	log.Infof("Decoding %v", body.Value.Payload)
	valueBuf, err := hex.DecodeString(body.Value.Payload)
	if err != nil {
		log.WithError(err).Error("fail to decode payload (hex)")
		return errors.Wrap(err, "fail to decode payload (hex)")
	}

	valueStr := string(valueBuf)

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

	// 1-6: Temp sxx.xx
	// 7-11: Hum xx.xx
	// 12-17: Lum xxx.xx
	// 18-21: Bat tension x.xx
	// 22-25: Sol tension x.xx
	// 26-31: Masse R1 xx.xxx
	// 32-37: Masse R2 xx.xxx
	// 38-43: Masse R3 xx.xxx
	// 44-49: Masse R4 xx.xxx
	// pour le futur, non implémenté
	// 50-54: TempR1 xx.xx
	// 55-59: TempR2 xx.xx
	// 60-64: TempR3 xx.xx
	// 65-69: TempR4 xx.xx

	values["temp"], err = strconv.ParseFloat(getValue(valueStr, 0, 6), 64)
	if err != nil {
		checkErr(ctx, err, "temp")
	}
	values["hum"], err = strconv.ParseFloat(getValue(valueStr, 6, 11), 64)
	if err != nil {
		checkErr(ctx, err, "hum")
	}
	values["lum"], err = strconv.ParseFloat(getValue(valueStr, 11, 17), 64)
	if err != nil {
		checkErr(ctx, err, "lum")
	}
	values["bat_tension"], err = strconv.ParseFloat(getValue(valueStr, 17, 21), 64)
	if err != nil {
		checkErr(ctx, err, "bat_tension")
	}
	values["sol_tension"], err = strconv.ParseFloat(getValue(valueStr, 21, 25), 64)
	if err != nil {
		checkErr(ctx, err, "sol_tension")
	}
	values["mass_r1"], err = strconv.ParseFloat(getValue(valueStr, 25, 31), 64)
	if err != nil {
		checkErr(ctx, err, "mass_r1")
	}
	values["mass_r2"], err = strconv.ParseFloat(getValue(valueStr, 31, 37), 64)
	if err != nil {
		checkErr(ctx, err, "mass_r2")
	}
	values["mass_r3"], err = strconv.ParseFloat(getValue(valueStr, 37, 43), 64)
	if err != nil {
		checkErr(ctx, err, "mass_r3")
	}
	values["mass_r4"], err = strconv.ParseFloat(getValue(valueStr, 43, 49), 64)
	if err != nil {
		checkErr(ctx, err, "mass_r4")
	}
	if len(valueStr) > 50 {
		values["temp_r1"], err = strconv.ParseFloat(getValue(valueStr, 49, 54), 64)
		if err != nil {
			checkErr(ctx, err, "temp_r1")
		}
		values["temp_r2"], err = strconv.ParseFloat(getValue(valueStr, 54, 59), 64)
		if err != nil {
			checkErr(ctx, err, "temp_r2")
		}
		values["temp_r3"], err = strconv.ParseFloat(getValue(valueStr, 59, 64), 64)
		if err != nil {
			checkErr(ctx, err, "temp_r3")
		}
		values["temp_r4"], err = strconv.ParseFloat(getValue(valueStr, 64, 69), 64)
		if err != nil {
			checkErr(ctx, err, "temp_r4")
		}
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

func checkErr(ctx context.Context, err error, value string) error {
	log := logger.Get(ctx)
	log.WithError(err).WithField("field", value).Error(err.Error())
	return err
}

func getValue(payload string, start, end int) string {
	return strings.Trim(payload[start:end], "\x00 ")
}
