package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// -----------------------------------------------------------------------------------------------

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NEXT_HOSTNAME", ""),
			},
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NEXT_API_KEY", ""),
			},
		},
		/*
		ResourcesMap: map[string]*schema.Resource{
			"customer": resourceItem(),
		},
		*/
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	hostname := d.Get("hostname").(string)
	api_key := d.Get("api_key").(int)
	return client.NewClient(hostname, api_key), nil
}

// -----------------------------------------------------------------------------------------------

type Client struct {
	hostname string
	api_key  string
	httpClient *http.Client
}

func NewClient(hostname string, api_key string) *Client {
	return &Client{
		hostname:   hostname,
		api_key:    api_key,
		httpClient: &http.Client{},
	}
}

// GetAll Retrieves all of the Items from the server
func (c *Client) GetAll() (*map[string]server.Item, error) {
	body, err := c.httpRequest("item", "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	items := map[string]server.Item{}
	err = json.NewDecoder(body).Decode(&items)
	if err != nil {
		return nil, err
	}
	return &items, nil
}

// GetItem gets an item with a specific name from the server
func (c *Client) GetItem(name string) (*server.Item, error) {
	body, err := c.httpRequest(fmt.Sprintf("item/%v", name), "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	item := &server.Item{}
	err = json.NewDecoder(body).Decode(item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// NewItem creates a new Item
func (c *Client) NewItem(item *server.Item) error {
	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(item)
	if err != nil {
		return err
	}
	_, err = c.httpRequest("item", "POST", buf)
	if err != nil {
		return err
	}
	return nil
}

// UpdateItem updates the values of an item
func (c *Client) UpdateItem(item *server.Item) error {
	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(item)
	if err != nil {
		return err
	}
	_, err = c.httpRequest(fmt.Sprintf("item/%s", item.Name), "PUT", buf)
	if err != nil {
		return err
	}
	return nil
}

// DeleteItem removes an item from the server
func (c *Client) DeleteItem(itemName string) error {
	_, err := c.httpRequest(fmt.Sprintf("item/%s", itemName), "DELETE", bytes.Buffer{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) httpRequest(path, method string, body bytes.Buffer) (closer io.ReadCloser, err error) {
	req, err := http.NewRequest(method, c.requestPath(path), &body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", c.authToken)
	switch method {
	case "GET":
	case "DELETE":
	default:
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		respBody := new(bytes.Buffer)
		_, err := respBody.ReadFrom(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("got a non 200 status code: %v", resp.StatusCode)
		}
		return nil, fmt.Errorf("got a non 200 status code: %v - %s", resp.StatusCode, respBody.String())
	}
	return resp.Body, nil
}

func (c *Client) requestPath(path string) string {
	return fmt.Sprintf("%s:%v/%s", c.hostname, c.port, path)
}

// -----------------------------------------------------------------------------------------------

/*
func resourceItem() *schema.Resource {
	fmt.Print()
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the resource, also acts as it's unique ID",
				ForceNew:     true,
				ValidateFunc: validateName,
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A description of an item",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "An optional list of tags, represented as a key, value pair",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
		Create: resourceCreateItem,
		Read:   resourceReadItem,
		Update: resourceUpdateItem,
		Delete: resourceDeleteItem,
		Exists: resourceExistsItem,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceCreateItem(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	tfTags := d.Get("tags").(*schema.Set).List()
	tags := make([]string, len(tfTags))
	for i, tfTag := range tfTags {
		tags[i] = tfTag.(string)
	}

	item := server.Item{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Tags:        tags,
	}

	err := apiClient.NewItem(&item)

	if err != nil {
		return err
	}
	d.SetId(item.Name)
	return nil
}

func resourceReadItem(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	itemId := d.Id()
	item, err := apiClient.GetItem(itemId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			d.SetId("")
		} else {
			return fmt.Errorf("error finding Item with ID %s", itemId)
		}
	}

	d.SetId(item.Name)
	d.Set("name", item.Name)
	d.Set("description", item.Description)
	d.Set("tags", item.Tags)
	return nil
}

func resourceUpdateItem(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	tfTags := d.Get("tags").(*schema.Set).List()
	tags := make([]string, len(tfTags))
	for i, tfTag := range tfTags {
		tags[i] = tfTag.(string)
	}

	item := server.Item{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Tags:        tags,
	}

	err := apiClient.UpdateItem(&item)
	if err != nil {
		return err
	}
	return nil
}

func resourceDeleteItem(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	itemId := d.Id()

	err := apiClient.DeleteItem(itemId)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceExistsItem(d *schema.ResourceData, m interface{}) (bool, error) {
	apiClient := m.(*client.Client)

	itemId := d.Id()
	_, err := apiClient.GetItem(itemId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
*/

// -----------------------------------------------------------------------------------------------
