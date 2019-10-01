package crm

import (
	"fmt"
	"strconv"

	"github.com/livechat/gokit/web/client"
)

type Customer struct {
	License License `json:"license"`
	Agents  []Agent `json:"agents"`
	Notes   []Note  `json:"notes"`
}

func (c *Customer) AgentByMail(mail string) Agent {
	var a Agent
	for _, i := range c.Agents {
		if i.Login == mail {
			return i
		}
	}

	return a
}

type License struct {
	ID              int    `json:"license_id"`
	Type            string `json:"type"`
	SalesPlan       string `json:"sales_plan"`
	Beta            string `json:"beta"`
	ProtocolVersion string `json:"lc_version"`
	Industry        string `json:"industry"`
	Notes           string `json:"notes"`
	CreatedAt       string `json:"creation_date"`
	ExpiresAt       string `json:"end_date"`
	Company         string `json:"company"`
}

type Agent struct {
	Name                  string
	Login                 string
	Role                  string `json:"acl"`
	Status                string `json:"status"`
	LoginStatus           string `json:"login_status"`
	Avatar                string `json:"avatar_path"`
	InactiveNotifications string `json:"inactive_notifications"`
	DesignVersion         string `json:"design_version"`
	ApiKey                string `json:"api_key"`
	Groups                []int  `json:"groups_member"`
}

type Note struct {
	Author, Body string
}

type Customers struct {
	host string
	http client.Caller
}

func (s Customers) Search(query string) (map[string]interface{}, error) {
	var r map[string]interface{}
	return r, s.http.Call(client.Option{
		URL:      fmt.Sprintf("%s/crm/customers/search?query=%s", s.host, query),
		Method:   "GET",
		Response: &r,
	})
}

//func (s Customers) AdvancedSearch(q Query) {
//	s.http.Call(client.Option{
//		URL: fmt.Sprintf(s.url, "/customers/advanced_search?query="+q["X"]),
//	})
//}

func (s Customers) ByLicense(number int) (Customer, error) {
	var r Customer
	return r, s.http.Call(client.Option{
		URL:      fmt.Sprintf("%s/crm/customers/%d", s.host, number),
		Response: &r,
	})
}

func (s Customers) Agent(license, email string) (Agent, error) {
	var a Agent
	var c Customer
	var l int
	var err error

	if l, err = strconv.Atoi(license); err != nil {
		return a, err
	}

	if c, err = s.ByLicense(l); err != nil {
		return a, err
	}

	return c.AgentByMail(email), nil
}
