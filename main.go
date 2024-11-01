package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	githubPatches "github.com/0x4f53/github-patches"
	gitlabPatches "github.com/0x4f53/gitlab-patches"
)

var signatures []Signature

func main() {

	err := makeDir(cacheDir)
	if err != nil {
		log.Println(err)
	}

	setFlags()

	logo()

	signatures = readSignatures()

	if *urlList != "" {
		urls, _ := readLines(*urlList)
		successfulUrls := fetchFromUrlList(urls)
		ScanFiles(successfulUrls)
		return
	}

	if *URL != "" {
		if !strings.HasPrefix(*URL, "http://") && !strings.HasPrefix(*URL, "https://") {
			log.Fatalf("Please enter a valid URL!")
		}
		var successfulUrls []string
		if *selenium {
			if !checkDockerInstalled() {
				log.Fatalf("Please install Docker to use Selenium mode!")
			}
			if !checkImageBuilt() {
				log.Fatalf("Attempting to build Selenium testing image from Dockerfile...")
				exec.Command("docker", "build", "-t", "selenium-integration", ".")
			}
			successfulUrls = []string{scrapeWithSelenium(*URL)}
		} else {
			successfulUrls = fetchFromUrlList([]string{*URL})
		}
		ScanFiles(successfulUrls)
		return
	}

	if *directory != "" {
		files, _ := getAllFiles(*directory)
		ScanFiles(files)
		return
	}

	if *file != "" {
		ScanFiles([]string{*file})
		return
	}

	if *github {
		githubPatches.GetCommitsInRange(githubPatches.GithubCacheDir, *from, *to, false)
		chunks, _ := listFiles(githubPatches.GithubCacheDir)

		var patches []string

		for _, chunk := range chunks {
			events, _ := githubPatches.ParseGitHubCommits(githubPatches.GithubCacheDir + chunk)

			for _, event := range events {
				for _, commit := range event.Payload.Commits {
					patches = append(patches, commit.PatchURL)
				}
			}

		}

		successfulUrls := fetchFromUrlList(patches)
		ScanFiles(successfulUrls)
		os.RemoveAll(githubPatches.GithubCacheDir)
		return
	}

	if *gitlab {
		commitData := gitlabPatches.GetGitlabCommits(100, 100)

		var patches []string
		for _, patch := range commitData {
			patches = append(patches, patch.CommitPatchURL)
		}

		successfulUrls := fetchFromUrlList(patches)
		ScanFiles(successfulUrls)
		os.RemoveAll(gitlabPatches.GitlabCacheDir)
		return
	}

	if *githubGists {
		gistData := githubPatches.GetLast100Gists()
		parsedGists, _ := githubPatches.ParseGistData(gistData)

		var gists []string
		for _, gist := range parsedGists {
			gists = append(gists, gist.RawURL)
		}

		successfulUrls := fetchFromUrlList(gists)
		ScanFiles(successfulUrls)
		return
	}

	if *phishtank {
		savePhishtankDataset()
		urls, _ := readLines(phishtankURLCache)
		successfulUrls := fetchFromUrlList(urls)
		ScanFiles(successfulUrls)
		return
	}

}
