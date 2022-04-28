package client

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"log"
	"net/http"
	"nexus-pusher/pkg/comps"
	"nexus-pusher/pkg/config"
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
	r2 string) []*comps.NexusComponent {

	var src, dst []*comps.NexusComponent
	wg1, wg2 := &sync.WaitGroup{}, &sync.WaitGroup{}
	tn := time.Now()

	wg1.Add(1)
	wg2.Add(1)
	go func() { src = s1.GetComponents(c1, nc1, r1); showFinalMessageForGetComponents(r1, src, tn); wg1.Done() }()
	go func() { dst = s2.GetComponents(c2, nc2, r2); showFinalMessageForGetComponents(r2, dst, tn); wg2.Done() }()
	wg1.Wait()
	wg2.Wait()

	return compareComponents(src, dst)
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
	cmpDiff := doCompareComponents(s1, c1, nc1, sc.SrcServerConfig.RepoName,
		s2, c2, nc2, sc.DstServerConfig.RepoName)
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
		srvUrl := fmt.Sprintf("%s%s%s?repository=%s", "http://127.0.0.1:8181",
			s2.BaseUrl,
			s2.ApiComponentsUrl,
			sc.DstServerConfig.RepoName)
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(data)
		if err != nil {
			log.Fatal(err)
		}
		body, err := s2.SendRequest(srvUrl, "POST", c2, &buf)
		if err != nil {
			log.Printf("%v", err)
		}
		fmt.Printf("%+v\n", body)

	} else {
		log.Printf("'%s' repo at server %s is in sync with repo '%s' at server %s, nothing to do.\n",
			sc.SrcServerConfig.RepoName,
			sc.SrcServerConfig.Server,
			sc.DstServerConfig.RepoName,
			sc.DstServerConfig.Server)
	}
}
