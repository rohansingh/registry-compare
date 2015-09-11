package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"sync"
)

type V1SearchResponse struct {
	Results []V1SearchResult
}

type V1SearchResult struct {
	Name string
}

func getV1Repos(uri string) []string {
	resp, _ := http.Get("http://" + uri + "/v1/search?q=")
	defer resp.Body.Close()

	var sr V1SearchResponse
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &sr)

	var list []string
	for _, r := range sr.Results {
		list = append(list, r.Name)
	}

	return list
}

func getV1Tags(uri, repo string) []string {
	var resp *http.Response
	for {
		var err error
		if resp, err = http.Get("http://" + uri + "/v1/repositories/" + repo + "/tags"); err == nil {
			defer resp.Body.Close()
			break
		}
	}

	var tags map[string]string
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &tags)

	var list []string
	for tag, _ := range tags {
		list = append(list, tag)
	}

	return list
}

type V2CatalogResponse struct {
	Repositories []string
}

type V2TagsResponse struct {
	Tags []string
}

func getV2Repos(uri string) []string {
	resp, _ := http.Get("http://" + uri + "/v2/_catalog?n=99999")
	defer resp.Body.Close()

	var r V2CatalogResponse
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &r)

	return r.Repositories
}

func getV2Tags(uri, repo string) []string {
	var resp *http.Response
	for {
		var err error
		if resp, err = http.Get("http://" + uri + "/v2/" + repo + "/tags/list?n=99999"); err == nil {
			defer resp.Body.Close()
			break
		}
	}

	var r V2TagsResponse
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &r)

	return r.Tags
}

type GetRepos func(string) []string
type GetTags func(string, string) []string

func getImages(uri string, reposFunc GetRepos, tagsFunc GetTags) []string {
	var images []string
	var wg sync.WaitGroup

	c := make(chan string)

	repos := reposFunc(uri)
	for _, r := range repos {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			for _, tag := range tagsFunc(uri, repo) {
				c <- repo + ":" + tag
			}
		}(r)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	for im := range c {
		images = append(images, im)
	}

	return images
}

func getMissingImages(source, target []string) []string {
	m := make(map[string]int, len(source))

	for _, im := range source {
		m[im] = 1
	}

	for _, im := range target {
		m[im] = m[im] - 1
	}

	var missing []string
	for im, v := range m {
		if v == 1 {
			missing = append(missing, im)
		}
	}

	sort.Strings(missing)
	return missing
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: ./registry-compare <V1_REGISTRY_ENDPOINT> <V2_REGISTRY_ENDPOINT>")
		os.Exit(1)
	}

	v1Uri, v2Uri := os.Args[1], os.Args[2]
	v1Images := getImages(v1Uri, getV1Repos, getV1Tags)
	v2Images := getImages(v2Uri, getV2Repos, getV2Tags)

	fmt.Fprintf(os.Stderr, "Images in %s that are missing in %s:\n", v1Uri, v2Uri)
	for _, im := range getMissingImages(v1Images, v2Images) {
		fmt.Println(im)
	}
}
