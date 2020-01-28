package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/log"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
)

func dummyPostSnapshotAtlanta(cmd m.CreateDashboardSnapshotCommand) (responseCode int) {

	responseCode = 200

	return
}

func getAltantaUrl() (url string) {

	empolisSection, err := setting.Raw.GetSection("empolis")

	if err != nil {
		log.Error(3, "getAltantaUrl: Unable to get url %v", err)
		return ""
	}

	url = empolisSection.Key("atlanta_url").MustString("")

	log.Info("atlantaUrl from config [%v]", url)

	return

}

func postSnapshotAtlanta(js *simplejson.Json, key string, deleteKey string) (responseCode int) {

	//postUrl := "http://host.docker.internal:8080/atlanta/iana/1/findings"

	var findingUUID string
	var theme string

	atlantaUrl := getAltantaUrl()

	templateArray := js.Get("templating").Get("list").MustArray()
	log.Info("templateArray len: [%v]", len(templateArray))

	for _, templateEntry := range templateArray {
		//log.Info("key[%v] value[%v]", k1, templateEntry)
		templateMap := templateEntry.(map[string]interface{})
		if templateMap["name"] == "finding_uuid" {
			log.Info("templateMap[\"current\"] [%v]", templateMap["current"])
			currentMap := templateMap["current"].(map[string]interface{})
			log.Info("finding_uuid found - value: [%v]", currentMap["value"])
			findingUUID = fmt.Sprintf("%v", currentMap["value"])
		}
	}

	if len(findingUUID) > 0 {
		log.Info("Empolis Dashboard Extension: findingUUID from Dashboard result '%v'", findingUUID)
	} else {
		log.Info("Empolis Dashboard Extension: findingUUID not set in Dashboard")
	}

	for _, templateEntry := range templateArray {
		//log.Info("key[%v] value[%v]", k1, templateEntry)
		templateMap := templateEntry.(map[string]interface{})
		if templateMap["name"] == "theme" {
			log.Info("templateMap[\"current\"] [%v]", templateMap["current"])
			currentMap := templateMap["current"].(map[string]interface{})
			log.Info("theme found - value: [%v]", currentMap["value"])
			theme = fmt.Sprintf("%v", currentMap["value"])
		}
	}

	if len(theme) > 0 {
		log.Info("Empolis Dashboard Extension: theme from Dashboard result '%v'", theme)
	} else {
		log.Info("Empolis Dashboard Extension: theme not set in Dashboard")
	}

	postData := map[string]interface{}{}

	postData["findingId"] = findingUUID
	postData["key"] = key
	postData["deleteKey"] = deleteKey
	postData["url"] = setting.ToAbsUrl("dashboard/snapshot/"+key) + "?kiosk" + "&theme=" + theme
	postData["deleteUrl"] = setting.ToAbsUrl("api/snapshots-delete/" + deleteKey)
	postData["type"] = "manual"

	mData, _ := json.Marshal(postData)
	println(string(mData))

	requestBody, err := json.Marshal(postData)

	if err != nil {
		log.Error(3, "postSnapshotAtlanta: Error marshalling postData %v", err)
	}

	postUrl := atlantaUrl + "/findings/" + findingUUID + "/snapshots"

	log.Info("Calculated postUrl [%v]", postUrl)

	response, err := http.Post(postUrl, "application/json", bytes.NewBuffer(requestBody))

	if err != nil {
		log.Error(3, "postSnapshotAtlanta: Error http post %v", err)
		responseCode = 500
		return
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Error(3, "postSnapshotAtlanta: Error Read response.body %v", err)
	}

	log.Info(string(body))

	responseCode = response.StatusCode

	return
}

func deleteSnapshotAtlanta(deleteKey string) (responseCode int) {

	atlantaUrl := getAltantaUrl()

	deleteUrl := atlantaUrl + "/findings/snapshots/" + deleteKey

	log.Info("Calculated deleteUrl [%v]", deleteUrl)

	request, err := http.NewRequest("DELETE", deleteUrl, nil)

	if err != nil {
		log.Error(3, "deleteSnapshotAtlanta: Error create Request %v", err)
	}

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		log.Error(3, "postSnapshotAtlanta: Error perform request %v", err)
	}

	responseCode = response.StatusCode

	return
}
