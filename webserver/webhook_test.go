package webserver

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func Test_Parser(t *testing.T) {
	payload, _ := hex.DecodeString("002008301100003a01150038f010e8044340e8")
	values := parsePayload(payload)

	fmt.Printf("%+v\n", values)
}
