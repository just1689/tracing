package tracing

import (
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
)

func httpPostJson(url string, r io.Reader) (err error) {
	resp, err := http.Post(url, "application/json", r)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode > 299 {
		logrus.Errorln(resp.StatusCode, err)
		var b []byte
		b, err = ioutil.ReadAll(resp.Body)
		logrus.Errorln(string(b))
	}
	return

}
