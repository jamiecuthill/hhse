package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"fmt"
)

var host string
var port int

const projectPath = "github.com/flypay/hhse"

func TestHhse(t *testing.T) {
	RegisterFailHandler(Fail)

	var service *gexec.Session

	BeforeSuite(func() {
		host = "localhost"
		port = 8000 + GinkgoParallelNode()

		packagePath, err := gexec.Build(projectPath)
		if err != nil {
			t.Fatal(err)
		}

		command := exec.Command(packagePath)
		command.Env = []string{
			fmt.Sprintf("PORT=%d", port),
		}

		service, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		if err != nil {
			t.Fatal(err)
		}

		Consistently(service).ShouldNot(gexec.Exit(), "service should run in the foreground but exited early")
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()

		if service != nil {
			service.Terminate().Wait()
		}
	})

	RunSpecs(t, "Hhse Suite")
}
