package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"fmt"
	"io/ioutil"
)

var _ = Describe("Hhse", func() {
	Describe("Healthcheck", func() {
		It("should respond OK", func() {
			resp, err := http.Get(fmt.Sprintf("http://%s:%d", host, port))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Describe("Menu", func() {
		It("should respond with menu", func() {
			resp, err := http.Get(fmt.Sprintf("http://%s:%d/menu", host, port))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(body)).To(MatchJSON(`{
				"items": [
					{ "name": "Stella" },
					{ "name": "Carlsberg" },
					{ "name": "Coors Light" },
					{ "name": "Carling" },
					{ "name": "Budweiser" }
				]
			}`))
		})
	})
})
