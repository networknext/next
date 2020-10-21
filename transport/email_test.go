package transport_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

// All tests listed below depend on test@networknext.com being a user in auth0
func TestMailChimpIntegrationStatusOK(t *testing.T) {
	HTTPServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	mailChimpHandler := transport.MailChimpHandler{
		HTTPHandler: *HTTPServer.Client(),
		MembersURI:  HTTPServer.URL,
	}
	t.Run("failure add - no email", func(t *testing.T) {
		err := mailChimpHandler.AddEmailToMailChimp("")
		assert.Error(t, err)
	})
	t.Run("failure tag - no email", func(t *testing.T) {
		err := mailChimpHandler.TagNewSignup("")
		assert.Error(t, err)
	})
	t.Run("success add", func(t *testing.T) {
		err := mailChimpHandler.AddEmailToMailChimp("test@test.com")
		assert.NoError(t, err)
	})
	t.Run("success tag", func(t *testing.T) {
		err := mailChimpHandler.TagNewSignup("test@test.com")
		assert.NoError(t, err)
	})
}

func TestMailChimpIntegrationStatusNotOK(t *testing.T) {
	HTTPServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}),
	)
	mailChimpHandler := transport.MailChimpHandler{
		HTTPHandler: *HTTPServer.Client(),
		MembersURI:  HTTPServer.URL,
	}
	t.Run("failure add", func(t *testing.T) {
		err := mailChimpHandler.AddEmailToMailChimp("test@test.com")
		assert.Error(t, err)
	})
	t.Run("failure tag", func(t *testing.T) {
		err := mailChimpHandler.TagNewSignup("test@test.com")
		assert.Error(t, err)
	})
}
