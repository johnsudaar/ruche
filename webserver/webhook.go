package webserver

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

func Webhook(resp http.ResponseWriter, req *http.Request, params map[string]string) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.Wrap(err, "fail to read request body")
	}
	fmt.Println(string(body))
	return nil
}
