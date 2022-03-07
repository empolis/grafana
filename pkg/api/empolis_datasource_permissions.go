package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	oauth "golang.org/x/oauth2/clientcredentials"
)

var datasourcePermissionsLogger = log.New("datasourcepermissions")

func init() {
	bus.AddHandler("sql", GetDataSourcesPermissions)
}

type IaApiClient struct {
	once   sync.Once
	client *http.Client
}

func (ac *IaApiClient) initClient() {
	ac.once.Do(func() {
		var oauthConfig = oauth.Config{
			ClientID:     setting.EmpolisIaApiClientId,
			ClientSecret: setting.EmpolisIaApiClientSecret,
			TokenURL:     setting.EmpolisIaApiTokenUrl,
		}

		ac.client = oauthConfig.Client(context.Background())
	})
}

func (ac *IaApiClient) getClient() *http.Client {
	ac.initClient()
	return ac.client
}

var iaApiOAuthClient IaApiClient

func getPermissionsFromIaApi(userId string, tenantId string) ([]string, error) {
	permissionsUrl := fmt.Sprintf("%s/permissions/%s/%s", setting.EmpolisIaApiUrl, tenantId, userId)
	req, _ := http.NewRequest("GET", permissionsUrl, nil)

	req.Header.Set("X-Tenant", tenantId)
	response, err := iaApiOAuthClient.getClient().Do(req)
	if err != nil {
		return nil, err
	}

	datasourcePermissionsLogger.Info(permissionsUrl)
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

	datasourcePermissionsLogger.Debug("Datasource permissions", "user role", query.User.OrgRole)
	if query.User.OrgRole == m.ROLE_ADMIN {
		// do not filter data sources
		query.Result = query.Datasources
		datasourcePermissionsLogger.Debug("Not filtering because of admin role", "ds length", len(query.Result))
		return nil
	}

	permissions, err := getPermissionsFromIaApi(query.User.Login, query.User.OrgName)

	if err != nil {
		datasourcePermissionsLogger.Error("Error getting IA API permissions", "error", err)
	}

	datasourcePermissionsLogger.Debug("Datasource permissions", "permissions", permissions)

	var filter []*regexp.Regexp

	for _, p := range permissions {
		r, _ := regexp.Compile(".*\\Q" + p + "\\E$")
		filter = append(filter, r)
	}

	datasourcePermissionsLogger.Debug("Filtering datasources", "ds filter", filter)
	for _, ds := range query.Datasources {
		datasourcePermissionsLogger.Debug("Inspecting datasource", "ds name", ds.Name)
		for _, regex := range filter {
			if regex.MatchString(ds.Name) {
				query.Result = append(query.Result, ds)
				break
			}
		}
	}
	return nil
}
