package mesoskafka

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	url        *url.URL
	httpClient *http.Client
}

func NewClient(hostname string, port int) Client {
	return NewClientForUrl(fmt.Sprintf("https://%s:%d", hostname, port))
}

func NewClientForUrl(rawurl string) Client {
	url, err := url.Parse(rawurl)

	if err != nil {
		panic(err)
	}

	c := Client{
		url:        url,
		httpClient: &http.Client{},
	}

	return c
}

func (c *Client) getFullUrl(apiEndpoint string) string {
	fullUrl, err := c.url.Parse(apiEndpoint)
	if err != nil {
		panic(err)
	}
	return fullUrl.String()
}

func (c *Client) getJson(apiEndpoint string) ([]byte, error) {
	resp, err := c.httpClient.Get(c.getFullUrl(apiEndpoint))
	if err != nil {
		return nil, fmt.Errorf("%s: %s", apiEndpoint, err)
	}
	if statusCodeErr := checkSuccessfullStatusCode(resp); statusCodeErr != nil {
		return nil, fmt.Errorf("%s: %s", apiEndpoint, statusCodeErr)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return body, nil
}

func (c *Client) putJson(apiEndpoint string, json []byte) ([]byte, error) {
	req, err := http.NewRequest("PUT", c.getFullUrl(apiEndpoint), bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if statusCodeErr := checkSuccessfullStatusCode(resp); statusCodeErr != nil {
		return nil, statusCodeErr
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return body, nil
}

func (c *Client) deleteJson(apiEndpoint string) ([]byte, error) {
	req, _ := http.NewRequest("DELETE", c.getFullUrl(apiEndpoint), nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if statusCodeErr := checkSuccessfullStatusCode(resp); statusCodeErr != nil {
		return nil, statusCodeErr
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return body, nil
}

func checkSuccessfullStatusCode(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		defer resp.Body.Close()
		responseBody := buf.String()
		return fmt.Errorf("Returned HTTP status: %s \nReturned HTTP body: %s", resp.Status, responseBody)
	}
	return nil
}

type Status struct {
	Brokers     []Broker `json:"brokers"`
}

type Broker struct {
	ID           string   `json:"id"`
	Active       bool     `json:"active"`
	Memory       int      `json:"mem"`
	Heap         int      `json:"heap"`
	Cpus         float64  `json:"cpus"`
	Log4jOptions string   `json:"log4jOptions"`
	Constraints  string   `json:"constraints"`
	JVMOptions   string   `json:"jvmOptions"`
	Options      string   `json:"options"`
	Failover     Failover `json:"failover"`
}

type Failover struct {
	Delay    string `json:"delay"`
	MaxDelay string `json:"maxDelay"`
	MaxTries int    `json:"maxTries"`
}

type Brokers struct {
	Brokers []Broker `json:"brokers"`
}

type MutateStatus struct {
	Status string `json:"started"`
}

type RebalanceStatus struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (c *Client) ApiBrokersStatus() (*Status, error) {
	body, e := c.getJson("/api/broker/list")

	if e != nil {
		return nil, e
	}

	var status Status
	e = json.Unmarshal(body, &status)
	if e != nil {
		return nil, e
	}

	return &status, nil
}

func queryStringFromBroker(broker *Broker) string {

	params := url.Values{}
	params.Add("broker", broker.ID)

	if broker.Cpus != 0 {
		params.Add("cpus", strconv.FormatFloat(broker.Cpus, 'f', 6, 64))
	}

	if broker.Memory != 0 {
		params.Add("mem", strconv.Itoa(broker.Memory))
	}

	if broker.Heap != 0 {
		params.Add("heap", strconv.Itoa(broker.Heap))
	}

	params.Add("jvmOptions", broker.JVMOptions)

	params.Add("log4jOptions", broker.Log4jOptions)

	params.Add("options", broker.Options)

	params.Add("constraints", broker.Constraints)

	if broker.Failover.Delay != "" {
		params.Add("failoverDelay", broker.Failover.Delay)
	}

	if broker.Failover.MaxDelay != "" {
		params.Add("failoverMaxDelay", broker.Failover.MaxDelay)
	}

	if broker.Failover.MaxTries != 0 {
		params.Add("failoverMaxTries", strconv.Itoa(broker.Failover.MaxTries))
	}

	return params.Encode()
}

func (c *Client) ApiBrokersAdd(broker *Broker) (*Brokers, error) {
	url := fmt.Sprintf("/api/broker/add?%s", queryStringFromBroker(broker))
	body, e := c.getJson(url)

	if e != nil {
		return nil, e
	}

	var response Brokers
	e = json.Unmarshal(body, &response)
	if e != nil {
		return nil, e
	}

	return &response, nil
}

func (c *Client) ApiBrokersStart(broker *Broker) (*MutateStatus, error) {
	url := fmt.Sprintf("/api/broker/start?broker=%s", broker.ID)
	body, e := c.getJson(url)

	if e != nil {
		return nil, e
	}

	var response MutateStatus
	e = json.Unmarshal(body, &response)
	if e != nil {
		return nil, e
	}

	return &response, nil
}

func (c *Client) ApiBrokersStop(BrokerId int) (*MutateStatus, error) {
	url := fmt.Sprintf("/api/broker/stop?broker=%d", BrokerId)
	body, e := c.getJson(url)

	if e != nil {
		return nil, e
	}

	var response MutateStatus
	e = json.Unmarshal(body, &response)
	if e != nil {
		return nil, e
	}

	return &response, nil
}

func (c *Client) ApiBrokersRemove(BrokerId int) (*MutateStatus, error) {
	url := fmt.Sprintf("/api/broker/remove?broker=%d", BrokerId)
	body, e := c.getJson(url)

	if e != nil {
		return nil, e
	}

	var response MutateStatus
	e = json.Unmarshal(body, &response)
	if e != nil {
		return nil, e
	}

	return &response, nil
}

func (c *Client) ApiBrokerUpdate(broker *Broker) (*MutateStatus, error) {
	url := fmt.Sprintf("/api/broker/update?%s", queryStringFromBroker(broker))
	body, e := c.getJson(url)

	if e != nil {
		return nil, e
	}

	var response MutateStatus
	e = json.Unmarshal(body, &response)
	if e != nil {
		return nil, e
	}

	return &response, nil
}

func (c *Client) ApiBrokersCreate(brokers *Brokers) error {

	for _, broker := range brokers.Brokers {
		_, err := c.ApiBrokersAdd(&broker)

		if err != nil {
			return err
		}

		_, err = c.ApiBrokersStart(&broker)

		if err != nil {
			return err
		}

		c.ApiBrokerRebalance()
	}

	return nil
}

func (c *Client) ApiBrokersDelete(BrokerIds []int) error {

	for _, brokerId := range BrokerIds {
		_, err := c.ApiBrokersStop(brokerId)

		if err != nil {
			return err
		}

		_, err = c.ApiBrokersRemove(brokerId)

		if err != nil {
			return err
		}

		c.ApiBrokerRebalance()
	}

	return nil
}

// This is a blocking call that returns when a rebalance is complete.
func (c *Client) ApiBrokerRebalance() error {
	url := "/api/broker/rebalance?broker=*"
	body, e := c.getJson(url)

	if e != nil {
		return e
	}

	var response MutateStatus
	e = json.Unmarshal(body, &response)
	if e != nil {
		return e
	}

	for {
		response, err := c.ApiBrokersRebalanceStatus()
		if err != err {
			return err
		}

		if response.Status == "idle" {
			break
		}

		fmt.Println("Waiting for rebalance...")
		time.Sleep(time.Second * 5)
	}

	return nil
}

func (c *Client) ApiBrokersRebalanceStatus() (*RebalanceStatus, error) {
	url := "/api/broker/rebalance"
	body, e := c.getJson(url)

	if e != nil {
		return nil, e
	}

	var response RebalanceStatus
	e = json.Unmarshal(body, &response)
	if e != nil {
		return nil, e
	}

	return &response, nil
}

func (c *Client) ApiBrokersUpdate(brokers []Broker) error {

	for _, broker := range brokers {
		id, err := strconv.Atoi(broker.ID)
		if err != nil {
			return err
		}

		_, err = c.ApiBrokersStop(id)
		if err != nil {
			return err
		}

		_, err = c.ApiBrokerUpdate(&broker)
		if err != nil {
			return err
		}

		_, err = c.ApiBrokersStart(&broker)
		if err != nil {
			return err
		}

		c.ApiBrokerRebalance()
	}

	return nil
}
