package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
)

var _ = Describe("Hhse", func() {
	Describe("Healthcheck", func() {
		It("should respond OK", func() {
			resp, err := http.Get(endpoint("/"))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Describe("Menu", func() {
		It("should respond with menu", func() {
			resp, err := http.Get(endpoint("/menu"))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header["Content-Type"]).To(ContainElement("application/json"))

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(body)).To(MatchJSON(`{
							"items": [
								{ "id": 1, "name": "Stella" },
								{ "id": 2, "name": "Carlsberg" },
								{ "id": 3, "name": "Coors Light" },
								{ "id": 4, "name": "Carling" },
								{ "id": 5, "name": "Budweiser" }
							]
						}`))
		})
	})

	Describe("Prices", func() {
		It("should start up with default prices", func() {
			resp, err := http.Get(endpoint("/prices"))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header["Content-Type"]).To(ContainElement("application/json"))

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(body)).To(MatchJSON(`{
							"prices": [
								{ "id": 1, "low": "£1.08", "high": "£1.08", "current": "£1.08", "trend": "" },
								{ "id": 2, "low": "£0.96", "high": "£0.96", "current": "£0.96", "trend": "" },
								{ "id": 3, "low": "£0.84", "high": "£0.84", "current": "£0.84", "trend": "" },
								{ "id": 4, "low": "£0.96", "high": "£0.96", "current": "£0.96", "trend": "" },
								{ "id": 5, "low": "£0.96", "high": "£0.96", "current": "£0.96", "trend": "" }
							]
						}`))
		})
	})
})

func endpoint(path string) string {
	path = strings.TrimLeft(path, "/")
	return fmt.Sprintf("http://%s:%d/%s", host, port, path)
}
