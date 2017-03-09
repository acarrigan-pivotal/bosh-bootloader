package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("director-address", func() {
	var (
		tempDirectory string
		args          []string
	)

	BeforeEach(func() {
		var err error

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		args = []string{
			"--state-dir", tempDirectory,
			"director-address",
		}
	})

	Context("when bbl manages the director", func() {
		BeforeEach(func() {
			state := []byte(`{
				"version": 3,
				"bosh": {
					"directorAddress": "some-director-url"
				}
			}`)
			err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the director address from the given state file", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("some-director-url"))
		})
	})

	Context("when bbl does not manage the director", func() {
		BeforeEach(func() {
			state := []byte(`{"version":3,"noDirector": true}`)
			err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a non zero exit code and prints a helpful error message", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err.Contents()).To(ContainSubstring("Error BBL does not manage this director."))
		})
	})

	Context("failure cases", func() {
		It("returns a non zero exit code when the bbl-state.json does not exist", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))

			expectedErrorMessage := fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)
			Expect(session.Err.Contents()).To(ContainSubstring(expectedErrorMessage))
		})

		It("returns a non zero exit code when the address does not exist", func() {
			state := []byte(`{"version":3}`)
			err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err.Contents()).To(ContainSubstring("Could not retrieve director address"))
		})
	})
})
