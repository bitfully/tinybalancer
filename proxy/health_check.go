// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proxy

import (
	"log"
	"time"

	"github.com/zehuamama/tinybalancer/util"
)

var HealthCheckTimeout = 5 * time.Second

// ReadAlive reads the alive status of the site
func (h *HTTPProxy) ReadAlive(url string) bool {
	h.RLock()
	defer h.RUnlock()
	return h.alive[url]
}

// SetAlive sets the alive status to the site
func (h *HTTPProxy) SetAlive(url string, alive bool) {
	h.Lock()
	defer h.Unlock()
	h.alive[url] = alive
}

// HealthCheck enable a health check goroutine for each agent
func (h *HTTPProxy) HealthCheck() {
	for host := range h.hostMap {
		go h.healthCheck(host)
	}
}

func (h *HTTPProxy) healthCheck(host string) {
	ticker := time.Tick(HealthCheckTimeout)
	for {
		select {
		case <-ticker:
			if !util.IsBackendAlive(host) && h.ReadAlive(host) {
				log.Printf("Site unreachable, remove %s from load balancer.", host)

				h.SetAlive(host, false)
				h.lb.Remove(host)
			} else if util.IsBackendAlive(host) && !h.ReadAlive(host) {
				log.Printf("Site reachable, add %s to load balancer.", host)

				h.SetAlive(host, true)
				h.lb.Add(host)
			}
		}
	}
}
