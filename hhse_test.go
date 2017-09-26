package main_test

import (
	. "github.com/flypay/hhse"
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

		It("should be cors enabled", func() {
			req, err := http.NewRequest(http.MethodOptions, endpoint("/"), nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Access-Control-Request-Method", "GET")
			req.Header.Set("Origin", "foo.com")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.Header["Access-Control-Allow-Origin"]).To(ContainElement("foo.com"))
			Expect(resp.Header["Access-Control-Allow-Methods"]).To(ContainElement(http.MethodGet))
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
				],
				"crash": null
			}`))
		})

		Context("when a bill event indicates a sale", func() {
			BeforeEach(func() {
				resp, err := http.Post(endpoint("/events"), "application/json", strings.NewReader(`{
					"checks": [{
						"checkNumber": "34",
						"createdAt": "2017-06-15 20:00:00",
						"lastUpdated": "2017-06-15 20:00:00"
					}],
					"bill": {
						"locationId": 123,
						"openedAt": "2017-06-15 20:00:00",
						"outstanding": 74.88,
						"lastUpdated": null,
						"closedAt": null,
						"discount": 0,
						"fullAmount": 2.16,
						"table": {
							"tableCode": "35",
							"guestCount": 2
						},
						"id": 1,
						"staff": [{
							"staffCode": "3",
							"name": null
						}],
						"serviceCharge": 0.00,
						"tipsPaid": 0,
						"products": [{
							"category": {
								"id": 1,
								"name": "drinks",
								"group": "drinks"
							},
							"price": 1.08,
							"priceSold": 0,
							"code": "42",
							"flypayProductId": 1,
							"productName": "Stella"
						}, {
							"category": {
								"id": 1,
								"name": "drinks",
								"group": "drinks"
							},
							"price": 1.08,
							"priceSold": 0,
							"code": "42",
							"flypayProductId": 1,
							"productName": "Stella"
						}],
						"payments": [],
						"type": "PayAtTable",
						"vat": 0.00
					}
				}`))
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
			})

			It("price of the purchased item should increase", func() {
				resp, err := http.Get(endpoint("/prices"))
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp.Header["Content-Type"]).To(ContainElement("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(body)).To(MatchJSON(`{
					"prices": [
						{ "id": 1, "low": "£1.08", "high": "£1.18", "current": "£1.18", "trend": "up" },
						{ "id": 2, "low": "£0.96", "high": "£0.96", "current": "£0.96", "trend": "" },
						{ "id": 3, "low": "£0.84", "high": "£0.84", "current": "£0.84", "trend": "" },
						{ "id": 4, "low": "£0.96", "high": "£0.96", "current": "£0.96", "trend": "" },
						{ "id": 5, "low": "£0.96", "high": "£0.96", "current": "£0.96", "trend": "" }
					],
					"crash": null
				}`))
			})
		})
	})

	Describe("Menu", func() {
		var product *Product

		BeforeEach(func() {
			product = NewProduct(1, "Beer", 100)
		})

		Describe("Product", func() {
			Describe("IncrPrice", func() {
				It("should increase the current price", func() {
					Expect(product.Current()).To(Equal(20))
					Expect(product.Low()).To(Equal(product.Current()))
					Expect(product.High()).To(Equal(product.Current()))

					product.IncrPrice()

					Expect(product.Current()).To(Equal(21))
					Expect(product.Trend).To(Equal("up"))
					Expect(product.High()).To(Equal(product.Current()))
				})

				It("should reset when reaching crash ratio", func() {
					Eventually(func() bool {
						product.IncrPrice()

						return product.Current() == product.Low()
					}).Should(BeTrue())

					Expect(product.Trend).To(Equal("down"))
				})
			})

			Describe("DecrPrice", func() {
				BeforeEach(func() {
					product.IncrPrice()
					product.IncrPrice()
					product.IncrPrice()
					product.IncrPrice()
				})

				It("should reduce the current price", func() {
					Expect(product.Low()).To(Equal(20))
					Expect(product.Current()).To(Equal(24))
					Expect(product.High()).To(Equal(product.Current()))

					product.DecrPrice()

					Expect(product.Low()).To(Equal(20))
					Expect(product.Current()).To(Equal(23))
					Expect(product.Trend).To(Equal("down"))
					Expect(product.High()).To(Equal(24))
				})

				It("should not reduce below lowest possible value", func() {
					product.DecrPrice()
					product.DecrPrice()
					product.DecrPrice()
					product.DecrPrice()
					product.DecrPrice()
					product.DecrPrice()
					product.DecrPrice()
					product.DecrPrice()
					product.DecrPrice()

					Expect(product.Low()).To(Equal(20))
					Expect(product.Current()).To(Equal(20))
					Expect(product.Trend).To(Equal(""))
					Expect(product.High()).To(Equal(24))
				})
			})
		})
	})
})

func endpoint(path string) string {
	path = strings.TrimLeft(path, "/")
	return fmt.Sprintf("http://%s:%d/%s", host, port, path)
}
