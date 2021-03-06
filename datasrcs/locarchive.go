// Copyright 2017 Jeff Foley. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package datasrcs

import (
	"context"
	"fmt"
	"time"

	"github.com/OWASP/Amass/v3/config"
	"github.com/OWASP/Amass/v3/eventbus"
	"github.com/OWASP/Amass/v3/requests"
	"github.com/OWASP/Amass/v3/systems"
)

// LoCArchive is the Service that handles access to the LoCArchive data source.
type LoCArchive struct {
	requests.BaseService

	SourceType string
	domain     string
	baseURL    string
}

// NewLoCArchive returns he object initialized, but not yet started.
func NewLoCArchive(sys systems.System) *LoCArchive {
	l := &LoCArchive{
		SourceType: requests.ARCHIVE,
		domain:     "webarchive.loc.gov",
		baseURL:    "http://webarchive.loc.gov/all",
	}

	l.BaseService = *requests.NewBaseService(l, "LoCArchive")
	return l
}

// Type implements the Service interface.
func (l *LoCArchive) Type() string {
	return l.SourceType
}

// OnStart implements the Service interface.
func (l *LoCArchive) OnStart() error {
	l.BaseService.OnStart()

	l.SetRateLimit(time.Second)
	return nil
}

// OnDNSRequest implements the Service interface.
func (l *LoCArchive) OnDNSRequest(ctx context.Context, req *requests.DNSRequest) {
	cfg := ctx.Value(requests.ContextConfig).(*config.Config)
	bus := ctx.Value(requests.ContextEventBus).(*eventbus.EventBus)
	if cfg == nil || bus == nil {
		return
	}

	if req.Name == "" || req.Domain == "" {
		return
	}

	if !cfg.IsDomainInScope(req.Name) {
		return
	}

	l.CheckRateLimit()

	names, err := crawl(ctx, l.baseURL, l.domain, req.Name, req.Domain)
	if err != nil {
		bus.Publish(requests.LogTopic, eventbus.PriorityHigh, fmt.Sprintf("%s: %v", l.String(), err))
		return
	}

	for _, name := range names {
		bus.Publish(requests.NewNameTopic, eventbus.PriorityHigh, &requests.DNSRequest{
			Name:   cleanName(name),
			Domain: req.Domain,
			Tag:    l.SourceType,
			Source: l.String(),
		})
	}
}
