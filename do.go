package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kjk/u"
)

const (
	// https://console.cloud.google.com/home/dashboard?folder=&organizationId=&project=cloudeval-255302
	gcpProject = "cloudeval-255302"
)

const (
	containerName = "code-eval-cont"
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

func buildDockerLocal(lang string) {
	dir := filepath.Join("docker", lang+"-local")
	imageName := "eval" + lang + ":latest"
	cmd := exec.Command("docker", "build", "-f", filepath.Join(dir, "Dockerfile"), "--tag", imageName, ".")
	u.RunCmdLoggedMust(cmd)
}

//for manually testing
func runDockerLocal(lang string) {
	panicIfNotValidLang(lang)
	imageName := "eval" + lang + ":latest"
	// we start only one container at a time, so we can use just one cname
	// (continer name)
	cmd := exec.Command("docker", "run", "--rm", "-it", "--name", containerName, "-p", "8533:8080", imageName)
	u.RunCmdLoggedMust(cmd)
}

func runUnitTests() {
	cmd := exec.Command("go", "test", "-v", "./...")
	u.RunCmdLoggedMust(cmd)
}

func startDockerLocal(lang string, verbose bool) func() {
	panicIfNotValidLang(lang)
	imageName := "eval" + lang + ":latest"
	// we start only one container at a time, so we can use just one cname
	// (continer name)
	cmd := exec.Command("docker", "run", "--rm", "--name", containerName, "-p", "8533:8080")
	if verbose {
		cmd.Args = append(cmd.Args, "--env", "VERBOSE=true")
	}
	cmd.Args = append(cmd.Args, imageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	must(err)
	return func() {
		// Maybe: use unique name for each container
		stopDockerContainer()
	}
}

func stopDockerContainer() {
	logf(context.Background(), "Stopping container '%s'\n", containerName)
	cmd := exec.Command("docker", "stop", containerName)
	err := cmd.Run()
	must(err)
}

func runEvalTestsLocal() {
	startDockerLocal("multi", true)
	defer stopDockerContainer()
	// TODO: write me
}
