package permissions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	oauth "golang.org/x/oauth2/clientcredentials"
)

var datasourcePermissionsLogger = log.New("datasourcepermissions")

var ErrNotImplemented = errors.New("not implemented")

type DatasourcePermissionsService interface {
	FilterDatasourcesBasedOnQueryPermissions(ctx context.Context, cmd *models.DatasourcesPermissionFilterQuery) error
}

func (hs *OSSDatasourcePermissionsService) permissionsFromIaApi(userId string, tenantId string) ([]string, error) {
	permissionsUrl := fmt.Sprintf("%s/permissions/%s/%s", setting.EmpolisIaApiUrl, tenantId, userId)
	req, _ := http.NewRequest("GET", permissionsUrl, nil)

	req.Header.Set("X-Tenant", tenantId)
	response, err := hs.iaApIClient.Do(req)
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

func (hs *OSSDatasourcePermissionsService) FilterDatasourcesBasedOnQueryPermissions(ctx context.Context, cmd *models.DatasourcesPermissionFilterQuery) error {
	if hs.iaApIClient == nil {
		return ErrNotImplemented
	}

	datasourcePermissionsLogger.Debug("Datasource permissions", "user role", cmd.User.OrgRole)
	if cmd.User.IsGrafanaAdmin || cmd.User.OrgRole == models.ROLE_ADMIN {
		// do not filter data sources
		datasourcePermissionsLogger.Debug("Not filtering because of admin role", "ds length", len(cmd.Result))
		return ErrNotImplemented
	}

	permissions, err := hs.permissionsFromIaApi(cmd.User.Login, cmd.User.OrgName)

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
	for _, ds := range cmd.Datasources {
		datasourcePermissionsLogger.Debug("Inspecting datasource", "ds name", ds.Name)
		for _, regex := range filter {
			if regex.MatchString(ds.Name) {
				cmd.Result = append(cmd.Result, ds)
				break
			}
		}
	}
	return nil
}

type OSSDatasourcePermissionsService struct {
	iaApIClient *http.Client
}

func ProvideDatasourcePermissionsService() *OSSDatasourcePermissionsService {
	if len(setting.EmpolisIaApiClientId) > 0 {
		oauthConfig := oauth.Config{
			ClientID:     setting.EmpolisIaApiClientId,
			ClientSecret: setting.EmpolisIaApiClientSecret,
			TokenURL:     setting.EmpolisIaApiTokenUrl,
		}
		return &OSSDatasourcePermissionsService{
			oauthConfig.Client(context.Background()),
		}
	}
	return &OSSDatasourcePermissionsService{}
}
