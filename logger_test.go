package lsego

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLogger(t *testing.T) {
	rb := defaultHTTPResponseBody{
		Code: 233,
		Msg:  "test",
	}

	buf := bytes.NewBuffer([]byte(""))
	log.SetOutput(buf)
	hs := httptest.NewServer(DefaultLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rbb, _ := json.Marshal(rb)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(rbb)
	})))

	resp, err := http.Get(hs.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	fmt.Println(buf.String())
	assert.Equal(t, fmt.Sprintf("err=%s)\n", rb.Msg), strings.Split(buf.String(), " ")[8])
}
