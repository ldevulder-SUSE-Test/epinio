package v1_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/epinio/epinio/acceptance/helpers/catalog"
	api "github.com/epinio/epinio/internal/api/v1"
	"github.com/epinio/epinio/pkg/api/core/v1/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Services API Application Endpoints, Mutations", func() {
	var namespace string
	const jsOK = `{"status":"ok"}`
	containerImageURL := "splatform/sample-app"

	BeforeEach(func() {
		namespace = catalog.NewNamespaceName()
		env.SetupAndTargetNamespace(namespace)
	})

	Describe("POST /api/v1/namespaces/:namespace/services/", func() {
		It("returns a 'bad request' for a non JSON body", func() {
			response, err := env.Curl("POST",
				fmt.Sprintf("%s%s/namespaces/%s/services",
					serverURL, api.Root, namespace),
				strings.NewReader(``))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(Equal("EOF"))
		})

		It("returns a 'bad request' for a non-object JSON body", func() {
			response, err := env.Curl("POST",
				fmt.Sprintf("%s%s/namespaces/%s/services",
					serverURL, api.Root, namespace),
				strings.NewReader(`[]`))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("json: cannot unmarshal array into Go value of type models.ServiceCreateRequest"))
		})

		It("returns a 'bad request' for JSON object without `name` key", func() {
			response, err := env.Curl("POST",
				fmt.Sprintf("%s%s/namespaces/%s/services",
					serverURL, api.Root, namespace),
				strings.NewReader(`{}`))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("Cannot create service without a name"))
		})

		It("returns a 'bad request' for JSON object empty `data` key", func() {
			response, err := env.Curl("POST",
				fmt.Sprintf("%s%s/namespaces/%s/services",
					serverURL, api.Root, namespace),
				strings.NewReader(`{
				    "name": "meh"
				}`))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("Cannot create service without data"))
		})

		It("returns a 'not found' when the namespace does not exist", func() {
			response, err := env.Curl("POST",
				fmt.Sprintf("%s%s/namespaces/bogus/services",
					serverURL, api.Root),
				strings.NewReader(`{
				    "name": "meh",
				    "data": {"host":"localhost", "port":"9999"}
				}`))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusNotFound), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("Targeted namespace 'bogus' does not exist"))
		})

		Context("with conflicting service", func() {
			var service string

			BeforeEach(func() {
				service = catalog.NewServiceName()
				env.MakeService(service)
			})

			AfterEach(func() {
				env.CleanupService(service)
			})

			It("returns a 'conflict'", func() {
				response, err := env.Curl("POST",
					fmt.Sprintf("%s%s/namespaces/%s/services",
						serverURL, api.Root, namespace),
					strings.NewReader(fmt.Sprintf(`{
					    "name": "%s",
					    "data": {"host":"localhost", "port":"9999"}
					}`, service)))
				Expect(err).ToNot(HaveOccurred())
				Expect(response).ToNot(BeNil())

				defer response.Body.Close()
				bodyBytes, err := ioutil.ReadAll(response.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusConflict), string(bodyBytes))
				var responseBody map[string][]errors.APIError
				json.Unmarshal(bodyBytes, &responseBody)
				Expect(responseBody["errors"][0].Title).To(
					Equal("Service '" + service + "' already exists"))
			})
		})

		Describe("Creation", func() {
			var service string

			BeforeEach(func() {
				service = catalog.NewServiceName()
			})

			AfterEach(func() {
				env.CleanupService(service)
			})

			It("creates the service", func() {
				response, err := env.Curl("POST",
					fmt.Sprintf("%s%s/namespaces/%s/services",
						serverURL, api.Root, namespace),
					strings.NewReader(fmt.Sprintf(`{
					    "name": "%s",
					    "data": {"host":"localhost", "port":"9999"}
					}`, service)))
				Expect(err).ToNot(HaveOccurred())
				Expect(response).ToNot(BeNil())

				defer response.Body.Close()
				bodyBytes, err := ioutil.ReadAll(response.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusCreated), string(bodyBytes))
				Expect(string(bodyBytes)).To(Equal(jsOK))
			})
		})
	})

	Describe("DELETE /api/v1/namespaces/:namespace/services/:service", func() {
		var service string

		BeforeEach(func() {
			service = catalog.NewServiceName()
		})

		It("returns a 'bad request' for a non JSON body", func() {
			response, err := env.Curl("DELETE",
				fmt.Sprintf("%s%s/namespaces/idontexist/services/%s",
					serverURL, api.Root, service),
				strings.NewReader(""))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(Equal("EOF"))
		})

		It("returns a 'bad request' for a non-object JSON body", func() {
			response, err := env.Curl("DELETE",
				fmt.Sprintf("%s%s/namespaces/idontexist/services/%s",
					serverURL, api.Root, service),
				strings.NewReader(`[]`))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("json: cannot unmarshal array into Go value of type models.ServiceDeleteRequest"))
		})

		It("returns a 'not found' when the namespace does not exist", func() {
			response, err := env.Curl("DELETE",
				fmt.Sprintf("%s%s/namespaces/idontexist/services/%s",
					serverURL, api.Root, service),
				strings.NewReader(`{ "unbind": false }`))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusNotFound), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("Targeted namespace 'idontexist' does not exist"))
		})

		It("returns a 'not found' when the service does not exist", func() {
			response, err := env.Curl("DELETE",
				fmt.Sprintf("%s%s/namespaces/%s/services/bogus",
					serverURL, api.Root, namespace),
				strings.NewReader(`{ "unbind": false }`))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusNotFound), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("Service 'bogus' does not exist"))

		})

		Context("with bound applications", func() {
			var app string
			var service string

			BeforeEach(func() {
				service = catalog.NewServiceName()
				app = catalog.NewAppName()
				env.MakeService(service)
				env.MakeContainerImageApp(app, 1, containerImageURL)
				env.BindAppService(app, service, namespace)
			})

			AfterEach(func() {
				env.CleanupApp(app)
				env.CleanupService(service)
			})

			It("returns 'bad request'", func() {
				response, err := env.Curl("DELETE",
					fmt.Sprintf("%s%s/namespaces/%s/services/%s",
						serverURL, api.Root, namespace, service),
					strings.NewReader(`{ "unbind": false }`))
				Expect(err).ToNot(HaveOccurred())
				Expect(response).ToNot(BeNil())

				defer response.Body.Close()
				bodyBytes, err := ioutil.ReadAll(response.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusBadRequest), string(bodyBytes))
				var responseBody map[string][]errors.APIError
				json.Unmarshal(bodyBytes, &responseBody)
				Expect(responseBody["errors"][0].Title).To(Equal("bound applications exist"))
				Expect(responseBody["errors"][0].Details).To(Equal(app))
			})

			It("unbinds and removes the service, when former is requested", func() {
				response, err := env.Curl("DELETE",
					fmt.Sprintf("%s%s/namespaces/%s/services/%s",
						serverURL, api.Root, namespace, service),
					strings.NewReader(`{ "unbind": true }`))
				Expect(err).ToNot(HaveOccurred())
				Expect(response).ToNot(BeNil())

				defer response.Body.Close()
				bodyBytes, err := ioutil.ReadAll(response.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK), string(bodyBytes))
				Expect(string(bodyBytes)).To(Equal("{\"boundapps\":[\"" + app + "\"]}"))
			})
		})

		Context("without bound applications", func() {
			var service string

			BeforeEach(func() {
				service = catalog.NewServiceName()
				env.MakeService(service)
			})

			It("removes the service", func() {
				response, err := env.Curl("DELETE",
					fmt.Sprintf("%s%s/namespaces/%s/services/%s",
						serverURL, api.Root, namespace, service),
					strings.NewReader(`{ "unbind" : false }`))
				Expect(err).ToNot(HaveOccurred())
				Expect(response).ToNot(BeNil())

				defer response.Body.Close()
				bodyBytes, err := ioutil.ReadAll(response.Body)
				Expect(response.StatusCode).To(Equal(http.StatusOK), string(bodyBytes))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(bodyBytes)).To(Equal("{\"boundapps\":[]}"))
			})
		})
	})

	Describe("POST /api/v1/namespaces/:namespace/applications/:arg/servicebindings/", func() {
		var app string

		BeforeEach(func() {
			app = catalog.NewAppName()
		})

		It("returns a 'bad request' for a non JSON body", func() {
			response, err := env.Curl("POST",
				fmt.Sprintf("%s%s/namespaces/%s/applications/%s/servicebindings",
					serverURL, api.Root, namespace, app),
				strings.NewReader(``))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(Equal("EOF"))
		})

		It("returns a 'bad request' for a non-object JSON body", func() {
			response, err := env.Curl("POST",
				fmt.Sprintf("%s%s/namespaces/%s/applications/%s/servicebindings",
					serverURL, api.Root, namespace, app),
				strings.NewReader(`[]`))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("json: cannot unmarshal array into Go value of type models.BindRequest"))
		})

		It("returns a 'bad request' for JSON object without `name` key", func() {
			response, err := env.Curl("POST",
				fmt.Sprintf("%s%s/namespaces/%s/applications/%s/servicebindings",
					serverURL, api.Root, namespace, app),
				strings.NewReader(`{}`))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("Cannot bind service without names"))
		})

		It("returns a 'not found' when the namespace does not exist", func() {
			response, err := env.Curl("POST",
				fmt.Sprintf("%s%s/namespaces/bogus/applications/_dummy_/servicebindings",
					serverURL, api.Root),
				strings.NewReader(`{ "names": ["meh"] }`))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusNotFound), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("Targeted namespace 'bogus' does not exist"))
		})

		It("returns a 'not found' when the application does not exist", func() {
			response, err := env.Curl("POST",
				fmt.Sprintf("%s%s/namespaces/%s/applications/bogus/servicebindings",
					serverURL, api.Root, namespace),
				strings.NewReader(`{ "names": ["meh"] }`))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusNotFound), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("Application 'bogus' does not exist"))
		})

		Context("with application", func() {
			var app string
			var service string

			BeforeEach(func() {
				app = catalog.NewAppName()
				service = catalog.NewServiceName()
				env.MakeContainerImageApp(app, 1, containerImageURL)
				env.MakeService(service)
			})

			AfterEach(func() {
				env.CleanupApp(app)
				env.CleanupService(service)
			})

			It("returns a 'not found' when the service does not exist", func() {
				response, err := env.Curl("POST",
					fmt.Sprintf("%s%s/namespaces/%s/applications/%s/servicebindings",
						serverURL, api.Root, namespace, app),
					strings.NewReader(`{ "names": ["bogus"] }`))
				Expect(err).ToNot(HaveOccurred())
				Expect(response).ToNot(BeNil())

				defer response.Body.Close()
				bodyBytes, err := ioutil.ReadAll(response.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusNotFound), string(bodyBytes))
				var responseBody map[string][]errors.APIError
				json.Unmarshal(bodyBytes, &responseBody)
				Expect(responseBody["errors"][0].Title).To(
					Equal("Service 'bogus' does not exist"))
			})

			Context("and already bound", func() {
				BeforeEach(func() {
					env.BindAppService(app, service, namespace)
				})

				It("returns a note about being bound", func() {
					response, err := env.Curl("POST",
						fmt.Sprintf("%s%s/namespaces/%s/applications/%s/servicebindings",
							serverURL, api.Root, namespace, app),
						strings.NewReader(fmt.Sprintf(`{ "names": ["%s"] }`, service)))
					Expect(err).ToNot(HaveOccurred())
					Expect(response).ToNot(BeNil())

					defer response.Body.Close()
					bodyBytes, err := ioutil.ReadAll(response.Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(response.StatusCode).To(Equal(http.StatusOK), string(bodyBytes))
					var responseBody map[string][]errors.APIError
					json.Unmarshal(bodyBytes, &responseBody)
					Expect(string(bodyBytes)).To(Equal(fmt.Sprintf(`{"wasbound":["%s"]}`, service)))
				})
			})

			It("binds the service", func() {
				response, err := env.Curl("POST",
					fmt.Sprintf("%s%s/namespaces/%s/applications/%s/servicebindings",
						serverURL, api.Root, namespace, app),
					strings.NewReader(fmt.Sprintf(`{ "names": ["%s"] }`, service)))
				Expect(err).ToNot(HaveOccurred())
				Expect(response).ToNot(BeNil())

				defer response.Body.Close()
				bodyBytes, err := ioutil.ReadAll(response.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK), string(bodyBytes))
				Expect(string(bodyBytes)).To(Equal(`{"wasbound":null}`))
			})
		})
	})

	Describe("DELETE /api/v1/namespaces/:namespace/applications/:app/servicebindings/:service", func() {
		var app string
		var service string

		BeforeEach(func() {
			service = catalog.NewServiceName()
			app = catalog.NewAppName()
		})

		It("returns a 'not found' when the namespace does not exist", func() {
			response, err := env.Curl("DELETE",
				fmt.Sprintf("%s%s/namespaces/idontexist/applications/%s/servicebindings/%s",
					serverURL, api.Root, app, service),
				strings.NewReader(""))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusNotFound), string(bodyBytes))

			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("Targeted namespace 'idontexist' does not exist"))
		})

		It("returns a 'not found' when the application does not exist", func() {
			response, err := env.Curl("DELETE",
				fmt.Sprintf("%s%s/namespaces/%s/applications/bogus/servicebindings/%s",
					serverURL, api.Root, namespace, service),
				strings.NewReader(""))
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			defer response.Body.Close()
			bodyBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusNotFound), string(bodyBytes))
			var responseBody map[string][]errors.APIError
			json.Unmarshal(bodyBytes, &responseBody)
			Expect(responseBody["errors"][0].Title).To(
				Equal("Application 'bogus' does not exist"))
		})

		Context("with application", func() {
			var app string

			BeforeEach(func() {
				app = catalog.NewAppName()
				env.MakeContainerImageApp(app, 1, containerImageURL)
			})

			AfterEach(func() {
				env.CleanupApp(app)
			})

			It("returns a 'not found' when the service does not exist", func() {
				response, err := env.Curl("DELETE",
					fmt.Sprintf("%s%s/namespaces/%s/applications/%s/servicebindings/bogus",
						serverURL, api.Root, namespace, app),
					strings.NewReader(""))
				Expect(err).ToNot(HaveOccurred())
				Expect(response).ToNot(BeNil())

				defer response.Body.Close()
				bodyBytes, err := ioutil.ReadAll(response.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusNotFound), string(bodyBytes))
				var responseBody map[string][]errors.APIError
				json.Unmarshal(bodyBytes, &responseBody)
				Expect(responseBody["errors"][0].Title).To(
					Equal("Service 'bogus' does not exist"))
			})

			Context("with service", func() {
				var service string

				BeforeEach(func() {
					service = catalog.NewServiceName()
					env.MakeService(service)
				})

				AfterEach(func() {
					env.CleanupService(service)
				})

				Context("already bound", func() {
					BeforeEach(func() {
						env.BindAppService(app, service, namespace)
					})

					It("unbinds the service", func() {
						response, err := env.Curl("DELETE",
							fmt.Sprintf("%s%s/namespaces/%s/applications/%s/servicebindings/%s",
								serverURL, api.Root, namespace, app, service),
							strings.NewReader(""))
						Expect(err).ToNot(HaveOccurred())
						Expect(response).ToNot(BeNil())

						defer response.Body.Close()
						bodyBytes, err := ioutil.ReadAll(response.Body)
						Expect(err).ToNot(HaveOccurred())
						Expect(response.StatusCode).To(Equal(http.StatusOK), string(bodyBytes))
						Expect(string(bodyBytes)).To(Equal(jsOK))
					})
				})

				It("returns a 'ok' even when the service is not bound (idempotency)", func() {
					response, err := env.Curl("DELETE",
						fmt.Sprintf("%s%s/namespaces/%s/applications/%s/servicebindings/%s",
							serverURL, api.Root, namespace, app, service),
						strings.NewReader(""))
					Expect(err).ToNot(HaveOccurred())
					Expect(response).ToNot(BeNil())

					defer response.Body.Close()
					bodyBytes, err := ioutil.ReadAll(response.Body)
					Expect(response.StatusCode).To(Equal(http.StatusOK), string(bodyBytes))
					Expect(string(bodyBytes)).To(Equal(jsOK))
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})
