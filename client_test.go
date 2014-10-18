package zeropush_test

import (
	. "github.com/sinangedik/zeropush"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sinangedik/zeropush/testutil"
	"net/http/httptest"
	"net/url"
)

var _ = Describe("Client", func() {

	var (
		client *Client
		server *httptest.Server
	)
	server = testutil.NewZeroTestServer()
	client = NewClient()
	server_url, _ := url.Parse(server.URL)
	client.BaseURL = server_url.String()

	Describe("/verify_credentials", func() {
		Context("With the correct credentials", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			response, err := client.VerifyCredentials()
			It("should verify the credentials", func() {
				Expect(err).Should(BeNil())
				Expect(response.Body[0]["message"]).To(Equal("authenticated"))
				Expect(response.Body[0]["auth_token_type"]).To(Equal("server_token"))
			})
		})
		Context("With incorrect credentials", func() {
			client.AuthToken = testutil.WRONG_AUTH_TOKEN
			response, err := client.VerifyCredentials()
			It("should fail verification of  the credentials", func() {
				Expect(err).ShouldNot(BeNil())
				Expect(response.Error["error"]).To(Equal("unauthorized"))
			})
		})
	})

	Describe("/inactive_tokens", func() {
		Context("With the correct credentials", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			response, err := client.GetInactiveTokens()
			It("should get the inactive tokents", func() {
				Expect(err).Should(BeNil())
				Expect(response.Body[0]["device_token"]).To(Equal("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcedf"))
				Expect(response.Body[0]["marked_inactive_at"]).To(Equal("2013-03-11T16:25:14-04:00"))
				Expect(response.Body[1]["device_token"]).To(Equal("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcedf"))
				Expect(response.Body[1]["marked_inactive_at"]).To(Equal("2013-03-11T16:25:14-04:00"))
			})
		})
		Context("With incorrect credentials", func() {
			client.AuthToken = testutil.WRONG_AUTH_TOKEN
			response, err := client.VerifyCredentials()
			It("should fail verification of  the credentials", func() {
				Expect(err).ShouldNot(BeNil())
				Expect(response.Error["error"]).To(Equal("unauthorized"))
			})
		})
	})
	Describe("/register", func() {
		Context("With only device token", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should successfully register the device", func() {
				response, err := client.Register("1236372819B36278G6783G21678321", "")
				Expect(err).Should(BeNil())
				Expect(response.Body[0]["message"]).To(Equal("ok"))
				Expect(response.GetHeader("X-Device-Quota")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Remaining")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Overage")).ShouldNot(Equal(""))
			})
			It("should successfully unregister the device", func() {
				response, err := client.Unregister("1236372819B36278G6783G21678321", "")
				Expect(err).Should(BeNil())
				Expect(response.Body[0]["message"]).To(Equal("ok"))
				Expect(response.GetHeader("X-Device-Quota")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Remaining")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Overage")).ShouldNot(Equal(""))
			})
		})
		Context("With no device token", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should fail to register the device", func() {
				response, err := client.Register("", "")
				Expect(err).ShouldNot(BeNil())
				Expect(response.Error["error"]).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Remaining")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Overage")).ShouldNot(Equal(""))
			})
			It("should fail to unregister the device", func() {
				response, err := client.Unregister("", "")
				Expect(err).ShouldNot(BeNil())
				Expect(response.Error["error"]).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Remaining")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Overage")).ShouldNot(Equal(""))
			})
		})
	})
	Describe("/subscribe", func() {
		Context("With device token and channel", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should successfully subscribe the device", func() {
				response, err := client.Subscribe("1236372819B36278G6783G21678321", "foo")
				Expect(err).Should(BeNil())
				Expect(response.Body[0]["device_token"]).To(Equal("1236372819B36278G6783G21678321"))
				Expect(len(response.Body[0]["channels"].([]interface{}))).To(Equal(1))
				Expect(response.GetHeader("X-Device-Quota")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Remaining")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Overage")).ShouldNot(Equal(""))
			})
			It("should successfully unsubscribe the device", func() {
				response, err := client.Unsubscribe("1236372819B36278G6783G21678321", "foo")
				Expect(err).Should(BeNil())
				Expect(response.Body[0]["device_token"]).To(Equal("1236372819B36278G6783G21678321"))
				Expect(len(response.Body[0]["channels"].([]interface{}))).To(Equal(0))
				Expect(response.GetHeader("X-Device-Quota")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Remaining")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Overage")).ShouldNot(Equal(""))
			})
		})
		Context("With no device token", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should fail to subscribe the device", func() {
				response, err := client.Subscribe("", "foo")
				Expect(err).ShouldNot(BeNil())
				Expect(response.Error["error"]).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Remaining")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Overage")).ShouldNot(Equal(""))
			})
			It("should fail to unsubscribe the device", func() {
				response, err := client.Unsubscribe("", "foo")
				Expect(err).ShouldNot(BeNil())
				Expect(response.Error["error"]).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Remaining")).ShouldNot(Equal(""))
				Expect(response.GetHeader("X-Device-Quota-Overage")).ShouldNot(Equal(""))
			})
		})

		Context("With no channel", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should fail to subscribe the device", func() {
				_, err := client.Subscribe("123423143214312431", "")
				Expect(err.Error()).Should(Equal("channel is not set"))
			})
			It("should fail to subscribe the device", func() {
				_, err := client.Unsubscribe("12321324324432", "")
				Expect(err.Error()).Should(Equal("channel is not set"))
			})
		})
	})
	Describe("/broadcast", func() {
		Context("With valid fields", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should send nitifications", func() {
				res, err := client.Broadcast("channel", "alert", "+1", "sound", "info", "10000", "true", "category")
				Expect(err).Should(BeNil())
				Expect(res.Body[0]["sent_count"]).Should(Equal(float64(100)))

			})
		})
	})
	Describe("/notify", func() {
		Context("With 2 device token, no alert, no info", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should come back with an error", func() {
				_, err := client.Notify("", "+1", "sound", "", "10000", "true", "category",
					"1234567891abcdef1234567890abcdef1234567890abcdef1234567890abcedf",
					"1234567890abcdef1234567890abcdef1234567890abcdef1234567890abceee")
				Expect(err).ShouldNot(BeNil())
			})
		})
		Context("With no device token", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should come back with an error", func() {
				_, err := client.Notify("alert", "+1", "sound", "info", "10000", "true", "category", "", "")
				Expect(err).ShouldNot(BeNil())
			})
		})
	})
	Describe("Get /devices/{device_token}", func() {
		Context("With no device token", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should come back with an error", func() {
				_, err := client.GetDevice("")
				Expect(err).ShouldNot(BeNil())
			})
		})
		Context("With a device token", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should succeed getting the device details", func() {
				res, err := client.GetDevice("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcedf")
				Expect(err).Should(BeNil())
				Expect(res.Body[0]["token"]).To(Equal("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcedf"))
				Expect(res.Body[0]["active"]).To(Equal(true))
				Expect(res.Body[0]["marked_inactive_at"]).To(BeNil())
				Expect(res.Body[0]["badge"]).To(Equal(float64(1)))
				Expect((res.Body[0]["channels"].([]interface{}))[0].(string)).To(Equal("testflight"))
			})
		})
	})
	Describe("/set_badge", func() {
		Context("With no device token", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should come back with an error", func() {
				_, err := client.SetBadge("", 5)
				Expect(err).ShouldNot(BeNil())
			})
		})
		Context("With invalid badge number", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should come back with an error", func() {
				_, err := client.SetBadge("", -1)
				Expect(err).ShouldNot(BeNil())
			})
		})
		Context("With a device token and valid badge", func() {
			client.AuthToken = testutil.CORRECT_AUTH_TOKEN
			It("should succeed setting the badge", func() {
				res, err := client.SetBadge("1234567891abcdef1234567890abcdef1234567890abcdef1234567890abcedf", 5)
				Expect(err).Should(BeNil())
				Expect(res.Body[0]["message"]).To(Equal("ok"))
			})
		})
	})
})
