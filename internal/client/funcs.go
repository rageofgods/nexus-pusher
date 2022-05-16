package client

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/goccy/go-json"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"nexus-pusher/internal/comps"
	"nexus-pusher/internal/config"
	"strings"
	"sync"
	"time"
)

func fileNameFromPath(path string) string { // Get last part of url chunk with filename information
	cmpPathSplit := strings.Split(path, "/")
	return strings.Trim(cmpPathSplit[len(cmpPathSplit)-1], "@")
}

// compareComponents will compare src to dst and return diff
func compareComponents(src []*comps.NexusComponent, dst []*comps.NexusComponent) []*comps.NexusComponent {
	// Make dst hash-map
	dstNca := make(map[string]struct{})
	for _, v := range dst {
		for _, vv := range v.Assets {
			dstNca[fileNameFromPath(vv.Path)] = struct{}{}
		}
	}

	// Search in dst
	var nc []*comps.NexusComponent
	for i, v := range src {
		var nca []*comps.NexusComponentAsset
		for ii, vv := range v.Assets {
			if _, ok := dstNca[fileNameFromPath(vv.Path)]; !ok {
				nca = append(nca, v.Assets[ii])
			}
		}
		if len(nca) != 0 {
			tmpSrc := src[i]
			tmpSrc.Assets = nca
			nc = append(nc, tmpSrc)
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
			showFinalMessageForGetComponents(r1, s1.Host, src, tn)
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
			showFinalMessageForGetComponents(r2, s1.Host, dst, tn)
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

func showFinalMessageForGetComponents(r string, s string, nc []*comps.NexusComponent, t time.Time) {
	log.Printf("Analyzing repo '%s' for server '%s' is done. Completed %d assets in %v.\n",
		r,
		s,
		len(nc),
		time.Since(t).Round(time.Second))
}

// RunNexusPusher client entry point
func RunNexusPusher(c *config.Client) {
	wg := &sync.WaitGroup{}
	for _, v := range c.SyncConfigs {
		wg.Add(1)
		syncConfig := v
		go func() { doSyncConfigs(c, syncConfig); wg.Done() }()
	}
	wg.Wait()
}

// ScheduleRunNexusPusher wrapper around RunNexusPusher to schedule syncs
func ScheduleRunNexusPusher(c *config.Client) error {
	loc, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		return err
	}

	s := gocron.NewScheduler(loc)
	j, err := s.Every(c.Daemon.SyncEveryMinutes).Minute().Do(RunNexusPusher, c)
	if err != nil {
		return fmt.Errorf("error: can't schedule sync. job: %v: error: %w", j, err)
	}
	s.StartBlocking()

	return nil
}

func doCheckServerStatus(server string) error {
	// Create URL for status checking
	srvUrl := fmt.Sprintf("%s%s%s", server, config.URIBase, config.URIStatus)
	// Define client
	c := comps.HttpClient()

	req, err := http.NewRequest("GET", srvUrl, nil)
	if err != nil {
		return err
	}
	// Send request to server
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: bad server status returned. server responded with: %s", resp.Status)
	}
	return nil
}

func doCheckRepoTypes(sc *config.SyncConfig) error {
	// Define variables
	s1 := comps.NewNexusServer(sc.SrcServerConfig.User, sc.SrcServerConfig.Pass,
		sc.SrcServerConfig.Server, config.URIBase, config.URIRepositories)
	s2 := comps.NewNexusServer(sc.DstServerConfig.User, sc.DstServerConfig.Pass,
		sc.DstServerConfig.Server, config.URIBase, config.URIRepositories)

	c1 := comps.HttpClient()
	c2 := comps.HttpClient()

	var nr1 []*comps.NexusRepository
	var nr2 []*comps.NexusRepository

	srvUrl1 := fmt.Sprintf("%s%s%s", s1.Host, s1.BaseUrl, s1.ApiComponentsUrl)
	srvUrl2 := fmt.Sprintf("%s%s%s", s2.Host, s2.BaseUrl, s2.ApiComponentsUrl)

	// Check repo for supported types
	if err := checkSupportedRepoTypes(comps.ComponentType(sc.Format)); err != nil {
		return err
	}

	// Creating error group for awaiting result from check repos types
	group := new(errgroup.Group)

	// Run first repo check
	group.Go(func() error {
		// Decode response 1
		b1, err := s1.SendRequest(srvUrl1, "GET", c1, nil)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b1, &nr1); err != nil {
			return err
		}

		for _, v := range nr1 {
			// Check if target repo is available on Nexus server
			if strings.EqualFold(v.Name, sc.SrcServerConfig.RepoName) {
				// Check for correct repo format
				if !strings.EqualFold(v.Format, sc.Format) {
					return fmt.Errorf("wrong repository '%s' format type for server %s. want: %s, get: %s",
						sc.SrcServerConfig.RepoName,
						sc.SrcServerConfig.Server,
						sc.Format,
						v.Format)
				}
				// If all ok, return
				return nil
			}
		}

		return fmt.Errorf("repo with name '%s' not found on server %s",
			sc.SrcServerConfig.RepoName, sc.SrcServerConfig.Server)
	})

	// Run second repo check
	group.Go(func() error {
		// Decode response 2
		b2, err := s2.SendRequest(srvUrl2, "GET", c2, nil)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b2, &nr2); err != nil {
			return err
		}

		for _, v := range nr2 {
			// Check if target repo is available on Nexus server
			if !strings.EqualFold(v.Name, sc.DstServerConfig.RepoName) {
				// Check for correct repo format
				if strings.EqualFold(v.Format, sc.Format) {
					return fmt.Errorf("wrong repository '%s' format type for server %s. want: %s, get: %s",
						sc.DstServerConfig.RepoName,
						sc.DstServerConfig.Server,
						sc.Format,
						v.Format)
				}
				// If all ok, return
				return nil
			}
		}

		return fmt.Errorf("repo with name '%s' not found on server %s",
			sc.DstServerConfig.RepoName, sc.DstServerConfig.Server)
	})

	// If we found error, return it
	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}

func doSyncConfigs(cc *config.Client, sc *config.SyncConfig) {
	// Define two groups of resources to compare remote repos
	s1 := comps.NewNexusServer(sc.SrcServerConfig.User, sc.SrcServerConfig.Pass,
		sc.SrcServerConfig.Server, config.URIBase, config.URIComponents)
	s2 := comps.NewNexusServer(sc.DstServerConfig.User, sc.DstServerConfig.Pass,
		sc.DstServerConfig.Server, config.URIBase, config.URIComponents)
	c1 := comps.HttpClient()
	c2 := comps.HttpClient()
	var nc1 []*comps.NexusComponent
	var nc2 []*comps.NexusComponent

	// Check repos type
	if err := doCheckRepoTypes(sc); err != nil {
		log.Printf("error: repository validation check failed: %v", err)
		return
	}

	// Check nexus-pusher server status
	if err := doCheckServerStatus(cc.Server); err != nil {
		log.Printf("error: server status check failed: %v", err)
		return
	}

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

func checkSupportedRepoTypes(repoType comps.ComponentType) error {
	switch repoType.Lower() {
	case comps.NPM:
		return nil
	case comps.PYPI:
		return nil
	default:
		return fmt.Errorf("error: unsuported component type %s", repoType)
	}
}
