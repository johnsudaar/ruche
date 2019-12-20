package webserver

import (
	"context"
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
	ctx := req.Context()
	log := logger.Get(ctx)
	config := config.Get()

	var body Input
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		log.WithError(err).Error("fail to decode body")
		return errors.Wrap(err, "fail to decode body")
	}

	valueBuf, err := hex.DecodeString(body.Value.Payload)
	if err != nil {
		log.WithError(err).Error("fail to decode payload (hex)")
		return errors.Wrap(err, "fail to decode payload (hex)")
	}

	valueStr := string(valueBuf)

	values := make(map[string]interface{})
	tags := make(map[string]string)
	values["location_alt"] = body.Location.Alt
	values["location_accuracy"] = body.Location.Accuracy
	values["location_lon"] = body.Location.Lon
	values["location_lat"] = body.Location.Lat

	if valueStr == "Restart" {
		log.Info("Restart: Ignoring...")
		return nil
	}
	if len(valueStr) <= 5 {
		log.Info("Model 1 decoding")
		// Model 1: Use only 1 value
		value, err := strconv.ParseFloat(valueStr, 32)
		if err != nil {
			log.WithError(err).Error("fail to decode payload (str)")
			return errors.Wrap(err, "fail to decode payload (str)")
		}
		values["temp"] = value
	} else {
		log.Info("Model 2 decoding")
		// 0-1: Rucher ID
		// 2-6: Temp xx.xx
		// 7-11: Hum xx.xx
		// 12-16: Lum xxxxx
		// 17-21: Bat tension
		// 22-26: Masse hivee1
		// 27-31: Masse hivee1
		// 32-36: Masse hivee1
		// 37-41: Masse hivee1
		hiveID, err := strconv.Atoi(valueStr[0:2])
		if err != nil {
			return checkErr(ctx, err)
		}
		temp, err := strconv.ParseFloat(valueStr[2:7], 64)
		if err != nil {
			return checkErr(ctx, err)
		}
		hum, err := strconv.ParseFloat(valueStr[7:12], 64)
		if err != nil {
			return checkErr(ctx, err)
		}
		lum, err := strconv.ParseFloat(valueStr[12:17], 64)
		if err != nil {
			return checkErr(ctx, err)
		}
		bat, err := strconv.ParseFloat(valueStr[17:22], 64)
		if err != nil {
			return checkErr(ctx, err)
		}
		hive1, err := strconv.ParseFloat(valueStr[22:27], 64)
		if err != nil {
			return checkErr(ctx, err)
		}
		hive2, err := strconv.ParseFloat(valueStr[27:32], 64)
		if err != nil {
			return checkErr(ctx, err)
		}
		hive3, err := strconv.ParseFloat(valueStr[32:37], 64)
		if err != nil {
			return checkErr(ctx, err)
		}
		hive4, err := strconv.ParseFloat(valueStr[37:41], 64)
		if err != nil {
			return checkErr(ctx, err)
		}

		values["hive_id"] = hiveID
		values["temp"] = temp
		values["hum"] = hum
		values["lum"] = lum
		values["bat"] = bat
		values["hive_1"] = hive1
		values["hive_2"] = hive2
		values["hive_3"] = hive3
		values["hive_4"] = hive4

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

func checkErr(ctx context.Context, err error) error {
	log := logger.Get(ctx)
	log.WithError(err).Error(err.Error())
	return err
}
