package beater

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/phenomenes/vago"
	"github.com/phenomenes/varnishstatbeat/config"
)

type Varnishstatbeat struct {
	done    chan struct{}
	client  publisher.Client
	varnish *vago.Varnish
	config  *vago.Config
	period  time.Duration
}

// New creates a new Beater
func New(b *beat.Beat, c *common.Config) (beat.Beater, error) {
	cfg := config.DefaultConfig
	if err := c.Unpack(&cfg); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	return &Varnishstatbeat{
		done: make(chan struct{}),
		config: &vago.Config{
			Path:    cfg.Path,
			Timeout: cfg.Timeout,
		},
		period: cfg.Period,
		client: b.Publisher.Connect(),
	}, nil
}

func (vb *Varnishstatbeat) Run(b *beat.Beat) error {
	var err error
	logp.Info("varnishstatbeat is running! Hit CTRL-C to stop it.")
	vb.varnish, err = vago.Open(vb.config)
	if err != nil {
		return err
	}
	defer vb.varnish.Close()

	ticker := time.NewTicker(vb.period)
	counter := 1
	for {
		select {
		case <-vb.done:
			return nil
		case <-ticker.C:
		}

		event := common.MapStr{
			"@timestamp": common.Time(time.Now()),
			"type":       "stats",
			"count":      counter,
		}

		for k, v := range vb.varnish.Stats() {
			key := strings.Replace(k, ".", "_", -1)
			if strings.Contains(key, "happy") {
				event[key] = strconv.FormatUint(v, 2)
			} else {
				event[key] = v
			}
		}

		vb.client.PublishEvent(event)
		logp.Info("Event sent")
		counter++
	}
}

func (vb *Varnishstatbeat) Stop() {
	vb.client.Close()
	close(vb.done)
}
