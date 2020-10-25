package main

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/kjk/u"
)

const (
	// https://console.cloud.google.com/home/dashboard?folder=&organizationId=&project=cloudeval-255302
	gcpProject = "cloudeval-255302"
)

var (
	validLangs = []string{"multi", "multi-local"}
)

func panicIfNotValidLang(lang string) {
	for _, s := range validLangs {
		if lang == s {
			return
		}
	}
	s := fmt.Sprintf("'%s' is not a supported language. Must be: %#v\n", lang, validLangs)
	panic(s)
}

func buildAndDeployToGCR(lang string) {
	panicIfNotValidLang(lang)

	// https://cloud.google.com/sdk/gcloud/reference/builds/submit
	{
		configPath := filepath.Join("docker", lang, "cloudbuild.yml")
		cmd := exec.Command("gcloud", "builds", "submit", "--project", gcpProject, "--config", configPath)
		u.RunCmdLoggedMust(cmd)
	}

	// https://cloud.google.com/sdk/gcloud/reference/beta/run/deploy
	{
		fullName := "eval-" + lang
		imageName := "gcr.io/cloudeval-255302/eval-" + lang

		cmd := exec.Command("gcloud", "beta", "run", "deploy", fullName, "--timeout=10m", "--project", gcpProject, "--image", imageName, "--platform", "managed", "--region", "us-central1", "--memory", "256Mi", "--allow-unauthenticated")
		u.RunCmdLoggedMust(cmd)
	}
}

func deployGcr() {
	buildAndDeployToGCR("multi")
}

func buildDockerLocal() {
	dir := filepath.Join("docker", "multi-local")
	cmd := exec.Command("docker", "build", "-f", filepath.Join(dir, "Dockerfile"), "-t", "eval-multi:latest", ".")
	u.RunCmdLoggedMust(cmd)
}
