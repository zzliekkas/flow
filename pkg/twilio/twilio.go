package twilio

import (
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type Client struct {
	client     *twilio.RestClient
	fromNumber string
}

func New(accountSid, authToken, fromNumber string) *Client {
	return &Client{
		client:     twilio.NewRestClientWithParams(twilio.ClientParams{Username: accountSid, Password: authToken}),
		fromNumber: fromNumber,
	}
}

// SendSMS 发送短信
func (c *Client) SendSMS(to, body string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(c.fromNumber)
	params.SetBody(body)
	_, err := c.client.Api.CreateMessage(params)
	return err
}
