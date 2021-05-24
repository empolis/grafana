package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	oauth "golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"
)

func init() {

	bus.AddHandler("sql", GetDataSourcesPermissions)
}

type AtlantaClient struct {
	once   sync.Once
	client *http.Client
}

func (ac *AtlantaClient) initClient() {
	ac.once.Do(func() {
		var oauthConfig = oauth.Config{
			ClientID:     setting.EmpolisAtlantaClientId,
			ClientSecret: setting.EmpolisAtlantaClientSecret,
			TokenURL:     setting.EmpolisAtlantaTokenUrl,
		}

		ac.client = oauthConfig.Client(context.Background())
	})
}

func (ac *AtlantaClient) getClient() *http.Client {
	ac.initClient()
	return ac.client
}

var atlantaOauthClient AtlantaClient

func getPermissionsAtlanta(userId string, tenantId string) ([]string, error) {
	atlantaUrl := setting.EmpolisAtlantaUrl
	permissionsUrl := fmt.Sprintf(atlantaUrl+"/permissions/%s/%s", tenantId, userId)
	req, _ := http.NewRequest("GET", permissionsUrl, nil)

	req.Header.Set("X-Tenant", tenantId)
	response, err := atlantaOauthClient.getClient().Do(req)
	if err != nil {
		return nil, err
	}

	log.Infof(permissionsUrl)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var permissions []string
	err = json.Unmarshal(body, &permissions)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func GetDataSourcesPermissions(query *m.DatasourcesPermissionFilterQuery) error {

	log.Debugf("### ", "user role", query.User.OrgRole)
	if query.User.OrgRole == m.ROLE_ADMIN {
		// do not filter data sources
		query.Result = query.Datasources
		log.Debugf("### do not filter data sources", "ds length", len(query.Result))
		return nil
	}

	permissions, err := getPermissionsAtlanta(query.User.Login, query.User.OrgName)

	if err != nil {
		log.Errorf(3, "### atlanta permissions", "error", err)
	}

	log.Debugf("###", "permissions", permissions)

	var filter []*regexp.Regexp

	for _, p := range permissions {
		r, _ := regexp.Compile(".*\\Q" + p + "\\E$")
		filter = append(filter, r)
	}

	log.Debugf("###", "ds filter", filter)
	for _, ds := range query.Datasources {
		log.Debugf("###", "ds name", ds.Name)
		for _, regex := range filter {
			if regex.MatchString(ds.Name) {
				query.Result = append(query.Result, ds)
				break
			}
		}
	}
	return nil
}
