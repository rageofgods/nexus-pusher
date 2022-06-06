package client

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"net/http"
	"nexus-pusher/internal/comps"
	"nexus-pusher/internal/config"
	http2 "nexus-pusher/pkg/http_clients"
	"nexus-pusher/pkg/utils"
	"strings"
	"sync"
	"time"
)

type client struct {
	config  *config.Client
	metrics *nexusClientMetrics
	version *comps.Version
}

func NewClient(version *comps.Version, config *config.Client, metrics *nexusClientMetrics) *client {
	return &client{config: config, metrics: metrics, version: version}
}

func fileNameFromPath(path string) string { // Get last part of url chunk with filename information
	cmpPathSplit := strings.Split(path, "/")
	return strings.Trim(cmpPathSplit[len(cmpPathSplit)-1], "@")
}

// compareComponents will compare src to dst and return diff
func compareComponents(src []*comps.NexusComponent, dst []*comps.NexusComponent) []*comps.NexusComponent {
	// Make dst hash-map
	dstNca := make(map[string]struct{}, len(dst))
	for _, v := range dst {
		for _, vv := range v.Assets {
			dstNca[strings.ToLower(vv.Path)] = struct{}{}
		}
	}

	// Search in dst
	var nc []*comps.NexusComponent
	for i, v := range src {
		var nca []*comps.NexusComponentAsset
		for ii, vv := range v.Assets {
			if _, ok := dstNca[strings.ToLower(vv.Path)]; !ok {
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

	ctx, cancel := context.WithCancel(context.Background())
	group, errCtx := errgroup.WithContext(ctx)
	var src, dst []*comps.NexusComponent
	tn := time.Now()

	group.Go(func() error {
		log.Infof("Start analyzing repository '%s' at server '%s'", r1, s1.Host)
		var err error
		src, err = s1.GetComponents(errCtx, c1, nc1, r1)
		if err != nil {
			cancel()
			return err
		} else {
			showFinalMessageForGetComponents(r1, s1.Host, src, tn)
		}
		return nil
	})
	group.Go(func() error {
		log.Infof("Start analyzing repository '%s' at server '%s'", r2, s2.Host)
		var err error
		dst, err = s2.GetComponents(errCtx, c2, nc2, r2)
		if err != nil {
			cancel()
			return err
		} else {
			showFinalMessageForGetComponents(r2, s2.Host, dst, tn)
		}
		return nil
	})

	// Check for errors in requests
	err := group.Wait()
	if err != nil {
		return nil, &utils.ContextError{
			Context: "doCompareComponents",
			Err: fmt.Errorf("unable to compare source repository '%s' at server '%s' "+
				"with destination repository '%s' at server '%s' beacuse of error: %v", r1, s1.Host, r2, s2.Host, err),
		}
	}

	return compareComponents(src, dst), nil
}

func showFinalMessageForGetComponents(repo string, server string, nc []*comps.NexusComponent, t time.Time) {
	log.Debugf("Analyzing repo '%s' for server '%s' is done. Completed %d assets in %v.",
		repo,
		server,
		len(nc),
		time.Since(t).Round(time.Second))
}

// RunNexusPusher client entry point
func (nc client) RunNexusPusher() {
	// Check nexus-pusher server status
	if err := nc.doCheckServerStatus(); err != nil {
		log.Errorf("server status check failed: %v", err)
		return
	}

	// Check server version
	if err := nc.doCheckServerVersion(); err != nil {
		log.Errorf("%v", err)
		return
	}

	wg := &sync.WaitGroup{}
	for _, v := range nc.config.SyncConfigs {
		wg.Add(1)
		go func(c *config.Client, syncConfig *config.SyncConfig) {
			nc.doSyncConfigs(c, syncConfig)
			wg.Done()
		}(nc.config, v)
	}
	wg.Wait()
}

// ScheduleRunNexusPusher wrapper around RunNexusPusher to schedule syncs
func (nc client) ScheduleRunNexusPusher(interval interface{}) error {
	loc, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		return fmt.Errorf("ScheduleRunNexusPusher: %w", err)
	}

	s := gocron.NewScheduler(loc)
	j, err := s.Every(interval).Minute().Do(nc.RunNexusPusher)
	if err != nil {
		return fmt.Errorf("can't schedule sync. job: %v: error: %w", j, err)
	}
	s.StartBlocking()

	return nil
}

func (nc client) doCheckServerStatus() error {
	// Create URL for status checking
	srvUrl := fmt.Sprintf("%s%s%s", nc.config.Server, config.URIBase, config.URIStatus)
	// Define client
	c := http2.HttpRetryClient()

	req, err := http.NewRequest("GET", srvUrl, nil)
	if err != nil {
		return fmt.Errorf("doCheckServerStatus: %w", err)
	}
	// Send request to server
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("doCheckServerStatus: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Export server status
		nc.metrics.staticMetrics.serverStatus.Set(0)

		return &utils.ContextError{
			Context: "doCheckServerStatus",
			Err:     fmt.Errorf("bad server status returned. server responded with: %s", resp.Status),
		}
	}

	// Export server status
	nc.metrics.staticMetrics.serverStatus.Set(1)

	return nil
}

func (nc client) doCheckServerVersion() error {
	// Create URL for status checking
	srvUrl := fmt.Sprintf("%s%s%s", nc.config.Server, config.URIBase, config.URIVersion)
	// Define client
	client := http2.HttpRetryClient()
	// Create request
	req, err := http.NewRequest("GET", srvUrl, nil)
	if err != nil {
		return fmt.Errorf("doCheckServerVersion: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	// Send request to server
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("doCheckServerVersion: %w", err)
	}
	// Read request body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("doCheckServerVersion: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &utils.ContextError{
			Context: "doCheckServerVersion",
			Err:     fmt.Errorf("bad server status returned. server responded with: %s", resp.Status),
		}
	}

	// Try to decode body to NexusExportComponents struct
	serverVersion := &comps.Version{}
	if err := json.Unmarshal(body, serverVersion); err != nil {
		return &utils.ContextError{
			Context: "doCheckServerVersion",
			Err:     fmt.Errorf("unable to validate server version: %w", err),
		}
	}

	// Export server version and build info
	nc.metrics.staticMetrics.serverInfo.WithLabelValues(serverVersion.Version, serverVersion.Build).Set(1)

	if nc.version.Version != serverVersion.Version {
		return &utils.ContextError{
			Context: "doCheckServerVersion",
			Err: fmt.Errorf("client version: '%s' is differ from server version: '%s'. please update",
				nc.version.Version, serverVersion.Version),
		}
	}

	return nil
}

func doCheckRepoTypes(sc *config.SyncConfig) error {
	// Define variables
	s1 := comps.NewNexusServer(sc.SrcServerConfig.User, sc.SrcServerConfig.Pass,
		sc.SrcServerConfig.Server, config.URIBase, config.URIRepositories)
	s2 := comps.NewNexusServer(sc.DstServerConfig.User, sc.DstServerConfig.Pass,
		sc.DstServerConfig.Server, config.URIBase, config.URIRepositories)

	c1 := http2.HttpRetryClient()
	c2 := http2.HttpRetryClient()

	var nr1 []*comps.NexusRepository
	var nr2 []*comps.NexusRepository

	srvUrl1 := fmt.Sprintf("%s%s%s", s1.Host, s1.BaseUrl, s1.ApiComponentsUrl)
	srvUrl2 := fmt.Sprintf("%s%s%s", s2.Host, s2.BaseUrl, s2.ApiComponentsUrl)

	// Check repo for supported types
	if err := checkSupportedRepoTypes(config.ComponentType(sc.Format)); err != nil {
		return fmt.Errorf("doCheckRepoTypes: %w", err)
	}

	// Creating error group for awaiting result from check repos types
	group := new(errgroup.Group)

	// Run first repo check
	group.Go(func() error {
		// Decode response 1
		b1, err := s1.SendRequest(srvUrl1, "GET", c1, nil)
		if err != nil {
			return fmt.Errorf("doCheckRepoTypes: %w", err)
		}
		if err := json.Unmarshal(b1, &nr1); err != nil {
			return fmt.Errorf("doCheckRepoTypes: %w", err)
		}

		for _, v := range nr1 {
			// Check if target repo is available on Nexus server
			if strings.EqualFold(v.Name, sc.SrcServerConfig.RepoName) {
				// Check for correct repo format
				if !strings.EqualFold(v.Format, sc.Format) {
					return &utils.ContextError{
						Context: "doCheckRepoTypes",
						Err: fmt.Errorf("wrong repository '%s' format type for server %s. want: %s, get: %s",
							sc.SrcServerConfig.RepoName,
							sc.SrcServerConfig.Server,
							sc.Format,
							v.Format),
					}
				}
				// If all ok, return
				return nil
			}
		}

		return &utils.ContextError{
			Context: "doCheckRepoTypes",
			Err: fmt.Errorf("repo with name '%s' not found on server %s",
				sc.SrcServerConfig.RepoName, sc.SrcServerConfig.Server),
		}
	})

	// Run second repo check
	group.Go(func() error {
		// Decode response 2
		b2, err := s2.SendRequest(srvUrl2, "GET", c2, nil)
		if err != nil {
			return fmt.Errorf("doCheckRepoTypes: %w", err)
		}
		if err := json.Unmarshal(b2, &nr2); err != nil {
			return fmt.Errorf("doCheckRepoTypes: %w", err)
		}

		for _, v := range nr2 {
			// Check if target repo is available on Nexus server
			if !strings.EqualFold(v.Name, sc.DstServerConfig.RepoName) {
				// Check for correct repo format
				if strings.EqualFold(v.Format, sc.Format) {
					return &utils.ContextError{
						Context: "doCheckRepoTypes",
						Err: fmt.Errorf("wrong repository '%s' format type for server %s. want: %s, get: %s",
							sc.DstServerConfig.RepoName,
							sc.DstServerConfig.Server,
							sc.Format,
							v.Format),
					}
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

func (nc client) doSyncConfigs(cc *config.Client, sc *config.SyncConfig) {
	// Define two groups of resources to compare remote repos
	s1 := comps.NewNexusServer(sc.SrcServerConfig.User, sc.SrcServerConfig.Pass,
		sc.SrcServerConfig.Server, config.URIBase, config.URIComponents)
	s2 := comps.NewNexusServer(sc.DstServerConfig.User, sc.DstServerConfig.Pass,
		sc.DstServerConfig.Server, config.URIBase, config.URIComponents)
	c1 := http2.HttpRetryClient()
	c2 := http2.HttpRetryClient()
	var nc1 []*comps.NexusComponent
	var nc2 []*comps.NexusComponent

	// Check repos type
	if err := doCheckRepoTypes(sc); err != nil {
		log.Errorf("repository validation check failed: %v", err)
		return
	}

	// Get repo diff
	cmpDiff, err := doCompareComponents(s1, c1, nc1, sc.SrcServerConfig.RepoName,
		s2, c2, nc2, sc.DstServerConfig.RepoName)
	if err != nil {
		log.Errorf("%v", err)
		return
	}

	// If we got some differences in two repos
	if len(cmpDiff) != 0 {
		log.Printf("Found %d differences between '%s' repo at server %s and '%s' repo at server %s:",
			len(cmpDiff),
			sc.SrcServerConfig.RepoName,
			sc.SrcServerConfig.Server,
			sc.DstServerConfig.RepoName,
			sc.DstServerConfig.Server)

		// Convert original nexus json to export type
		data := genNexExpCompFromNexComp(sc.ArtifactsSource, cmpDiff)
		data.NexusServer = comps.NexusServer{
			Host:             sc.DstServerConfig.Server,
			BaseUrl:          config.URIBase,
			ApiComponentsUrl: config.URIComponents,
			Username:         sc.DstServerConfig.User,
			Password:         sc.DstServerConfig.Pass,
		}

		// Send diff data to nexus-pusher server
		pc := newPushClient(cc.Server, cc.ServerAuth.User, cc.ServerAuth.Pass, nc.metrics)

		// Use basic auth to get JWT token
		if err := pc.authorize(); err != nil {
			log.Errorf("%v", err)
			return
		}

		// Send compare request to nexus-pusher server
		body, err := pc.sendComparedRequest(data, sc.DstServerConfig.RepoName)
		if err != nil {
			log.Errorf("%v", err)
			return
		}

		// Start server polling to get request results
		if err := pc.pollComparedResults(body, sc.DstServerConfig.RepoName, sc.DstServerConfig.Server); err != nil {
			log.Errorf("%v", err)
		}
	} else {
		// Log repo is 'in-sync' event
		log.Printf("'%s' repo at server %s is in sync with repo '%s' at server %s, nothing to do.",
			sc.SrcServerConfig.RepoName,
			sc.SrcServerConfig.Server,
			sc.DstServerConfig.RepoName,
			sc.DstServerConfig.Server)
	}
}

func checkSupportedRepoTypes(repoType config.ComponentType) error {
	switch repoType.Lower() {
	case config.NPM:
		return nil
	case config.PYPI:
		return nil
	case config.MAVEN2:
		return nil
	case config.NUGET:
		return nil
	default:
		return &utils.ContextError{
			Context: "checkSupportedRepoTypes",
			Err:     fmt.Errorf("unsuported component type %s", repoType),
		}
	}
}
