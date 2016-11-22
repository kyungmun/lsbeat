package beater

import (
	"fmt"
	"time"
	"io/ioutil"
	"path/filepath"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/kyungmun/lsbeat/config"
)

type Lsbeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
	period time.Duration
	path   string
	lastIndexTime time.Time
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Lsbeat{
		done: make(chan struct{}),
		config: config,
	}
	return bt, nil
}

func (bt *Lsbeat) Run(b *beat.Beat) error {
	logp.Info("lsbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	counter := 1
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		//event := common.MapStr{
		//	"@timestamp": common.Time(time.Now()),
		//	"type":       b.Name,
		//	"counter":    counter,
		//}
		//bt.client.PublishEvent(event)

		bt.listDir(bt.config.Path, b.Name)
		bt.lastIndexTime = time.Now()
		logp.Info("Event sent")
		counter++
	}
}

func (bt *Lsbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}


func (bt *Lsbeat) listDir(dirFile string, beatname string) {
    files, _ := ioutil.ReadDir(dirFile)
    for _, f := range files {
        t := f.ModTime()
	path := filepath.Join(dirFile, f.Name())
	
        if t.After(bt.lastIndexTime) {
            // index all files and directories on init
		 event := common.MapStr{
	            "@timestamp": common.Time(time.Now()),
	            "type":       beatname,
	            "modtime":    common.Time(t),
	            "filename":   f.Name(),
	            "path":       path,
	            "directory":  f.IsDir(),
	            "filesize":   f.Size(),
	        }
		logp.Info("Event : %s", event)
        	//bt.client.PublishEvent(event) 
        }

        if f.IsDir() {
            bt.listDir(path, beatname)
        }
    }
}
