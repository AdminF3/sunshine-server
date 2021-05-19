package cmd

import (
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-systemd/v22/daemon"
)

// watchdog sends "WATCHDOG=1" once every few seconds (depending on systemd
// configuration) when running as systemd service in order to notify for its
// aliveness. If something goes wrong with the HTTP server or the database and
// systemd doesn't receive a signal for some time it will try to restart the service.
//
// When NOT running as systemd service or systemd's watchdog is not enabled,
// this function is a noop.
func watchdog(url string) {
	daemon.SdNotify(false, daemon.SdNotifyReady)
	interval, err := daemon.SdWatchdogEnabled(false)
	if err != nil || interval == 0 {
		return
	}

	for {
		_, err := http.Get(url)
		if err == nil {
			daemon.SdNotify(false, daemon.SdNotifyWatchdog)
		} else {
			log.Println("watchdog:", err)
		}

		time.Sleep(interval / 3)
	}
}
