package cmd

import (
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	prom "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	Exporters            []string
	ExportersHTTPTimeout int
	ExportersHostList    map[string]string
	ExportersHostAlias   string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"RequestURI": r.RequestURI,
		"UserAgent":  r.UserAgent(),
	}).Debug("handling new request")
	h.Merge(w)
}

func (h Handler) Merge(w io.Writer) {
	mfs := map[string]*prom.MetricFamily{}

	responses := make([]map[string]*prom.MetricFamily, 10)
	responsesMu := sync.Mutex{}
	httpClientTimeout := time.Second * time.Duration(h.ExportersHTTPTimeout)

	var wg sync.WaitGroup
	for _, url := range h.Exporters {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			log.WithField("url", u).Debug("Getting remote metrics")

			httpClient := http.Client{Timeout: httpClientTimeout}
			resp, err := httpClient.Get(u)
			if err != nil {
				log.WithField("url", u).Errorf("HTTP connection failed: %v", err)
				return
			}

			defer resp.Body.Close()

			tp := new(expfmt.TextParser)
			part, err := tp.TextToMetricFamilies(resp.Body)
			if err != nil {
				log.WithField("url", u).Errorf("Parse response body to metrics: %v", err)
				return
			}

			if host, ok := h.ExportersHostList[u]; ok {
				for _, mf := range part {
					for i := 0; i < len(mf.Metric); i++ {
						mf.Metric[i].Label = append(mf.Metric[i].Label, &prom.LabelPair{Name: &h.ExportersHostAlias, Value: &host})
					}
				}
			}

			responsesMu.Lock()
			responses = append(responses, part)
			responsesMu.Unlock()
		}(url)
	}

	wg.Wait()

	for _, part := range responses {
		for n, mf := range part {
			mfo, ok := mfs[n]
			if ok {
				mfo.Metric = append(mfo.Metric, mf.Metric...)
			} else {
				mfs[n] = mf
			}
		}
	}

	names := []string{}
	for n := range mfs {
		names = append(names, n)
	}
	sort.Strings(names)

	enc := expfmt.NewEncoder(w, expfmt.FmtText)
	for _, n := range names {
		err := enc.Encode(mfs[n])
		if err != nil {
			log.Error(err)
			return
		}
	}
}
