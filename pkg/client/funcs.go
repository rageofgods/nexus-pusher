package client

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron"
	"log"
	"net/http"
	"nexus-pusher/pkg/comps"
	"nexus-pusher/pkg/config"
	"strings"
	"sync"
	"time"
)

// compareComponents will compare src to dst and return diff
func compareComponents(src []*comps.NexusComponent, dst []*comps.NexusComponent) []*comps.NexusComponent {
	// Make dst hash-map
	s := make(map[string]*comps.NexusComponent)
	for i, v := range dst {
		s[fmt.Sprintf("%s-%s", v.Name, v.Version)] = dst[i]
	}
	// Search in dst
	var nc []*comps.NexusComponent
	for i, v := range src {
		if _, ok := s[fmt.Sprintf("%s-%s", v.Name, v.Version)]; !ok {
			nc = append(nc, src[i])
		}
	}
	return nc
}

func doCompareComponents(s1 *comps.NexusServer,
	c1 *http.Client,
	nc1 []*comps.NexusComponent,
	r1 string,
	s2 *comps.NexusServer,
	c2 *http.Client,
	nc2 []*comps.NexusComponent,
	r2 string) ([]*comps.NexusComponent, error) {

	var src, dst []*comps.NexusComponent
	wg := &sync.WaitGroup{}
	var isError bool
	tn := time.Now()
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(2)
	go func() {
		var err error
		src, err = s1.GetComponents(ctx, c1, nc1, r1)
		if err != nil {
			cancel()
			log.Printf("%v", err)
			isError = true
		} else {
			showFinalMessageForGetComponents(r1, src, tn)
		}
		wg.Done()
	}()
	go func() {
		var err error
		dst, err = s2.GetComponents(ctx, c2, nc2, r2)
		if err != nil {
			cancel()
			log.Printf("%v", err)
			isError = true
		} else {
			showFinalMessageForGetComponents(r2, dst, tn)
		}
		wg.Done()
	}()
	wg.Wait()
	// Check for errors in requests
	if isError {
		return nil, fmt.Errorf("error: unable to compare repositories")
	}

	return compareComponents(src, dst), nil
}

func showFinalMessageForGetComponents(r string, nc []*comps.NexusComponent, t time.Time) {
	log.Printf("Analyzing repo '%s' is done. Completed %d assets in %v.\n",
		r,
		len(nc),
		time.Since(t).Round(time.Second))
}

// RunNexusPusher client entry point
func RunNexusPusher(c *config.NexusConfig) {
	wg := &sync.WaitGroup{}
	for _, v := range c.Client.SyncConfigs {
		wg.Add(1)
		syncConfig := v
		go func() { doSyncConfigs(&c.Client, syncConfig); wg.Done() }()
	}
	wg.Wait()
}

// ScheduleRunNexusPusher wrapper around RunNexusPusher to schedule syncs
func ScheduleRunNexusPusher(c *config.NexusConfig) error {
	loc, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		return err
	}

	s := gocron.NewScheduler(loc)
	j, err := s.Every(c.Client.Daemon.SyncEveryMinutes).Minute().Do(RunNexusPusher, c)
	if err != nil {
		return fmt.Errorf("error: can't schedule sync. job: %v: error: %v", j, err)
	}
	s.StartBlocking()

	return nil
}

func doSyncConfigs(cc *config.Client, sc *config.SyncConfig) {
	s1 := comps.NewNexusServer(sc.SrcServerConfig.User, sc.SrcServerConfig.Pass,
		sc.SrcServerConfig.Server, config.URIBase, config.URIComponents)
	s2 := comps.NewNexusServer(sc.DstServerConfig.User, sc.DstServerConfig.Pass,
		sc.DstServerConfig.Server, config.URIBase, config.URIComponents)
	c1 := comps.HttpClient()
	c2 := comps.HttpClient()
	var nc1 []*comps.NexusComponent
	var nc2 []*comps.NexusComponent
	// Get repo diff
	cmpDiff, err := doCompareComponents(s1, c1, nc1, sc.SrcServerConfig.RepoName,
		s2, c2, nc2, sc.DstServerConfig.RepoName)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	// If we got some differences in two repos
	if len(cmpDiff) != 0 {
		log.Printf("Found %d differences between '%s' repo at server %s and '%s' repo at server %s:\n",
			len(cmpDiff),
			sc.SrcServerConfig.RepoName,
			sc.SrcServerConfig.Server,
			sc.DstServerConfig.RepoName,
			sc.DstServerConfig.Server)
		for _, v := range cmpDiff {
			for _, vv := range v.Assets {
				log.Printf("Component name: %s, Version: %s, Asset: %s\n",
					v.Name,
					v.Version,
					componentNameFromPath(vv.Path))
			}
		}
		// Convert original nexus json to export type
		data := genNexExpCompFromNexComp(cmpDiff)
		data.NexusServer = comps.NexusServer{
			Host:             sc.DstServerConfig.Server,
			BaseUrl:          config.URIBase,
			ApiComponentsUrl: config.URIComponents,
			Username:         sc.DstServerConfig.User,
			Password:         sc.DstServerConfig.Pass,
		}

		// Send diff data to nexus-pusher server
		pc := newPushClient(cc.Server, cc.ServerAuth.User, cc.ServerAuth.Pass)
		// Use basic auth to get JWT token
		if err := pc.authorize(); err != nil {
			log.Printf("%v", err)
			return
		}
		// Send compare request to nexus-pusher server
		body, err := pc.sendComparedRequest(data, sc.DstServerConfig.RepoName)
		if err != nil {
			log.Printf("%v", err)
			return
		}
		// Start server polling to get request results
		if err := pc.pollComparedResults(body); err != nil {
			log.Printf("%v", err)
		}
	} else {
		log.Printf("'%s' repo at server %s is in sync with repo '%s' at server %s, nothing to do.\n",
			sc.SrcServerConfig.RepoName,
			sc.SrcServerConfig.Server,
			sc.DstServerConfig.RepoName,
			sc.DstServerConfig.Server)
	}
}

func componentNameFromPath(cmpPath string) string {
	cmpPathSplit := strings.Split(cmpPath, "/")
	return cmpPathSplit[len(cmpPathSplit)-1]
}
