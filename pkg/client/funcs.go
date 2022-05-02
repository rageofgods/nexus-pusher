package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"log"
	"net/http"
	"nexus-pusher/pkg/comps"
	"nexus-pusher/pkg/config"
	"nexus-pusher/pkg/server"
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

func RunNexusPusher(c *config.NexusConfig) {
	wg := &sync.WaitGroup{}
	for _, v := range c.Client.SyncConfigs {
		wg.Add(1)
		value := v
		go func() { doSyncConfigs(value); wg.Done() }()
	}
	wg.Wait()
}

func doSyncConfigs(sc *config.SyncConfig) {
	s1 := comps.NewNexusServer(sc.SrcServerConfig.User, sc.SrcServerConfig.Pass,
		sc.SrcServerConfig.Server, baseUrl, apiComponentsUrl)
	s2 := comps.NewNexusServer(sc.DstServerConfig.User, sc.DstServerConfig.Pass,
		sc.DstServerConfig.Server, baseUrl, apiComponentsUrl)
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
			BaseUrl:          baseUrl,
			ApiComponentsUrl: apiComponentsUrl,
			Username:         sc.DstServerConfig.User,
			Password:         sc.DstServerConfig.Pass,
		}
		// Send diff data to nexus-pusher server
		log.Printf("Sending components diff to %s server...", s2.Host)
		if sc.PushTo == "" {
			sc.PushTo = sc.DstServerConfig.Server
		}
		srvUrl := fmt.Sprintf("%s%s%s?repository=%s", sc.PushTo,
			s2.BaseUrl,
			s2.ApiComponentsUrl,
			sc.DstServerConfig.RepoName)
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(data)
		if err != nil {
			log.Printf("%v", err)
			return
		}
		body, err := s2.SendRequest(srvUrl, "POST", c2, &buf)
		if err != nil {
			log.Printf("%v", err)
			return
		}
		log.Printf("Sending components diff to %s succesfully complete.", s2.Host)
		// Get results from server
		getRequestResult(body, s2, c2, sc.PushTo)
	} else {
		log.Printf("'%s' repo at server %s is in sync with repo '%s' at server %s, nothing to do.\n",
			sc.SrcServerConfig.RepoName,
			sc.SrcServerConfig.Server,
			sc.DstServerConfig.RepoName,
			sc.DstServerConfig.Server)
	}
}

func getRequestResult(body []byte, s *comps.NexusServer, c *http.Client, pushTo string) {
	// Convert body to Message type
	msg := &server.Message{}
	if err := json.Unmarshal(body, msg); err != nil {
		log.Printf("%v", err)
		return
	}
	log.Printf("Starting server polling for message id %s to get upload results...", msg.ID)
	// Queue http polling
	srvUrl := fmt.Sprintf("%s%s%s?uuid=%s", pushTo,
		s.BaseUrl,
		s.ApiComponentsUrl,
		msg.ID)

	// Poll maximum for 1800 seconds (30 min)
	limitTime := 1800
	for x := 1; x < limitTime; x++ {
		body, err := s.SendRequest(srvUrl, "GET", c, nil)
		if err != nil {
			log.Printf("%v", err)
			return
		}
		if err := json.Unmarshal(body, msg); err != nil {
			log.Printf("%v", err)
			return
		}
		if msg.Complete {
			log.Printf("Server polling for message id %s is complete with response from server:\n>>>\n%v<<<\n",
				msg.ID,
				msg.Response)
			return
		}
		if x%30 == 0 {
			log.Printf("Server polling for message id %s in progress... %d seconds passed",
				msg.ID,
				x)
		}
		time.Sleep(1 * time.Second)
	}
	// Show error if we don't get results in time
	log.Printf("error: unable to get results from for message id %s in %d seconds",
		msg.ID,
		limitTime)
}
