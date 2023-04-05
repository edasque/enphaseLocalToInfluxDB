package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/antchfx/htmlquery"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	influxclient "github.com/influxdata/influxdb1-client/v2"
)

const configFilePath = "config.yaml"

type JWTToken struct {
	Token          string `json:"token"`
	ExpiresAt      int    `json:"expires_at"`
	GenerationTime int    `json:"generation_time"`
}

type Inverters []struct {
	Serialnumber    string `json:"serialNumber"`
	Lastreportdate  int    `json:"lastReportDate"`
	Devtype         int    `json:"devType"`
	Lastreportwatts int    `json:"lastReportWatts"`
	Maxreportwatts  int    `json:"maxReportWatts"`
}

type enphaseMetrics struct {
	Production  []Production  `json:"production"`
	Consumption []Consumption `json:"consumption"`
	// Storage     []Storage     `json:"storage"`
}
type Lines struct {
	WNow             float64 `json:"wNow"`
	WhLifetime       float64 `json:"whLifetime"`
	VarhLeadLifetime float64 `json:"varhLeadLifetime"`
	VarhLagLifetime  float64 `json:"varhLagLifetime"`
	VahLifetime      float64 `json:"vahLifetime"`
	RmsCurrent       float64 `json:"rmsCurrent"`
	RmsVoltage       float64 `json:"rmsVoltage"`
	ReactPwr         float64 `json:"reactPwr"`
	ApprntPwr        float64 `json:"apprntPwr"`
	PwrFactor        float64 `json:"pwrFactor"`
	WhToday          float64 `json:"whToday"`
	WhLastSevenDays  float64 `json:"whLastSevenDays"`
	VahToday         float64 `json:"vahToday"`
	VarhLeadToday    float64 `json:"varhLeadToday"`
	VarhLagToday     float64 `json:"varhLagToday"`
}
type Production struct {
	Type             string  `json:"type"`
	ActiveCount      int     `json:"activeCount"`
	ReadingTime      int     `json:"readingTime"`
	WNow             float64 `json:"wNow"`
	WhLifetime       float64 `json:"whLifetime"`
	MeasurementType  string  `json:"measurementType,omitempty"`
	VarhLeadLifetime float64 `json:"varhLeadLifetime,omitempty"`
	VarhLagLifetime  float64 `json:"varhLagLifetime,omitempty"`
	VahLifetime      float64 `json:"vahLifetime,omitempty"`
	RmsCurrent       float64 `json:"rmsCurrent,omitempty"`
	RmsVoltage       float64 `json:"rmsVoltage,omitempty"`
	ReactPwr         float64 `json:"reactPwr,omitempty"`
	ApprntPwr        float64 `json:"apprntPwr,omitempty"`
	PwrFactor        float64 `json:"pwrFactor,omitempty"`
	WhToday          float64 `json:"whToday,omitempty"`
	WhLastSevenDays  float64 `json:"whLastSevenDays,omitempty"`
	VahToday         float64 `json:"vahToday,omitempty"`
	VarhLeadToday    float64 `json:"varhLeadToday,omitempty"`
	VarhLagToday     float64 `json:"varhLagToday,omitempty"`
	Lines            []Lines `json:"lines,omitempty"`
}

type Consumption struct {
	Type             string  `json:"type"`
	ActiveCount      int     `json:"activeCount"`
	MeasurementType  string  `json:"measurementType"`
	ReadingTime      int     `json:"readingTime"`
	WNow             float64 `json:"wNow"`
	WhLifetime       float64 `json:"whLifetime"`
	VarhLeadLifetime float64 `json:"varhLeadLifetime"`
	VarhLagLifetime  float64 `json:"varhLagLifetime"`
	VahLifetime      float64 `json:"vahLifetime"`
	RmsCurrent       float64 `json:"rmsCurrent"`
	RmsVoltage       float64 `json:"rmsVoltage"`
	ReactPwr         float64 `json:"reactPwr"`
	ApprntPwr        float64 `json:"apprntPwr"`
	PwrFactor        float64 `json:"pwrFactor"`
	WhToday          float64 `json:"whToday"`
	WhLastSevenDays  float64 `json:"whLastSevenDays"`
	VahToday         float64 `json:"vahToday"`
	VarhLeadToday    float64 `json:"varhLeadToday"`
	VarhLagToday     float64 `json:"varhLagToday"`
	Lines            []Lines `json:"lines"`
}

// type Storage struct {
// 	Type        string `json:"type"`
// 	ActiveCount int    `json:"activeCount"`
// 	ReadingTime int    `json:"readingTime"`
// 	WNow        int    `json:"wNow"`
// 	WhNow       int    `json:"whNow"`
// 	State       string `json:"state"`
// }

type SenseTrends struct {
	// Steps       int       `json:"steps"`
	// Start       time.Time `json:"start"`
	// End         time.Time `json:"end"`
	Consumption struct {
		Total float64 `json:"total"`
		// 	Totals  []float64 `json:"totals"`
		// 	Devices []struct {
		// 		ID   string `json:"id"`
		// 		Name string `json:"name"`
		// 		Icon string `json:"icon"`
		// 		Tags struct {
		// 			DefaultUserDeviceType       string `json:"DefaultUserDeviceType"`
		// 			DeviceListAllowed           string `json:"DeviceListAllowed"`
		// 			TimelineAllowed             string `json:"TimelineAllowed"`
		// 			UserDeleted                 string `json:"UserDeleted"`
		// 			UserDeviceTypeDisplayString string `json:"UserDeviceTypeDisplayString"`
		// 			UserEditable                string `json:"UserEditable"`
		// 			UserMergeable               string `json:"UserMergeable"`
		// 			UserShowBubble              string `json:"UserShowBubble"`
		// 			UserShowInDeviceList        string `json:"UserShowInDeviceList"`
		// 		} `json:"tags,omitempty"`
		// 		History     []float64 `json:"history"`
		// 		Avgw        float64   `json:"avgw"`
		// 		TotalKwh    float64   `json:"total_kwh"`
		// 		TotalCost   int       `json:"total_cost"`
		// 		Pct         float64   `json:"pct"`
		// 		CostHistory []int     `json:"cost_history"`
		// 		Tags0       struct {
		// 			Alertable             string    `json:"Alertable"`
		// 			AlwaysOn              string    `json:"AlwaysOn"`
		// 			DateCreated           time.Time `json:"DateCreated"`
		// 			DateFirstUsage        string    `json:"DateFirstUsage"`
		// 			DefaultUserDeviceType string    `json:"DefaultUserDeviceType"`
		// 			DeployToMonitor       string    `json:"DeployToMonitor"`
		// 			DeviceListAllowed     string    `json:"DeviceListAllowed"`
		// 			ModelCreatedVersion   string    `json:"ModelCreatedVersion"`
		// 			ModelUpdatedVersion   string    `json:"ModelUpdatedVersion"`
		// 			NameUseredit          string    `json:"name_useredit"`
		// 			OriginalName          string    `json:"OriginalName"`
		// 			PeerNames             []struct {
		// 				Name                        string  `json:"Name"`
		// 				UserDeviceType              string  `json:"UserDeviceType"`
		// 				Percent                     float64 `json:"Percent"`
		// 				Icon                        string  `json:"Icon"`
		// 				UserDeviceTypeDisplayString string  `json:"UserDeviceTypeDisplayString"`
		// 			} `json:"PeerNames"`
		// 			Pending                     string `json:"Pending"`
		// 			Revoked                     string `json:"Revoked"`
		// 			TimelineAllowed             string `json:"TimelineAllowed"`
		// 			TimelineDefault             string `json:"TimelineDefault"`
		// 			Type                        string `json:"Type"`
		// 			UserDeletable               string `json:"UserDeletable"`
		// 			UserDeviceTypeDisplayString string `json:"UserDeviceTypeDisplayString"`
		// 			UserEditable                string `json:"UserEditable"`
		// 			UserEditableMeta            string `json:"UserEditableMeta"`
		// 			UserMergeable               string `json:"UserMergeable"`
		// 			UserShowBubble              string `json:"UserShowBubble"`
		// 			UserShowInDeviceList        string `json:"UserShowInDeviceList"`
		// 		} `json:"tags,omitempty"`
		// 		Tags1 struct {
		// 			Alertable             string    `json:"Alertable"`
		// 			AlwaysOn              string    `json:"AlwaysOn"`
		// 			DateCreated           time.Time `json:"DateCreated"`
		// 			DateFirstUsage        string    `json:"DateFirstUsage"`
		// 			DefaultUserDeviceType string    `json:"DefaultUserDeviceType"`
		// 			DeployToMonitor       string    `json:"DeployToMonitor"`
		// 			DeviceListAllowed     string    `json:"DeviceListAllowed"`
		// 			ModelCreatedVersion   string    `json:"ModelCreatedVersion"`
		// 			ModelUpdatedVersion   string    `json:"ModelUpdatedVersion"`
		// 			NameUseredit          string    `json:"name_useredit"`
		// 			OriginalName          string    `json:"OriginalName"`
		// 			PeerNames             []struct {
		// 				Name                        string  `json:"Name"`
		// 				UserDeviceType              string  `json:"UserDeviceType"`
		// 				Percent                     float64 `json:"Percent"`
		// 				Icon                        string  `json:"Icon"`
		// 				UserDeviceTypeDisplayString string  `json:"UserDeviceTypeDisplayString"`
		// 			} `json:"PeerNames"`
		// 			Pending                     string `json:"Pending"`
		// 			Revoked                     string `json:"Revoked"`
		// 			TimelineAllowed             string `json:"TimelineAllowed"`
		// 			TimelineDefault             string `json:"TimelineDefault"`
		// 			Type                        string `json:"Type"`
		// 			UserDeletable               string `json:"UserDeletable"`
		// 			UserDeviceTypeDisplayString string `json:"UserDeviceTypeDisplayString"`
		// 			UserEditable                string `json:"UserEditable"`
		// 			UserEditableMeta            string `json:"UserEditableMeta"`
		// 			UserMergeable               string `json:"UserMergeable"`
		// 			UserShowBubble              string `json:"UserShowBubble"`
		// 			UserShowInDeviceList        string `json:"UserShowInDeviceList"`
		// 		} `json:"tags,omitempty"`
		// 		Tags2 struct {
		// 			Alertable             string    `json:"Alertable"`
		// 			AlwaysOn              string    `json:"AlwaysOn"`
		// 			DateCreated           time.Time `json:"DateCreated"`
		// 			DateFirstUsage        string    `json:"DateFirstUsage"`
		// 			DefaultUserDeviceType string    `json:"DefaultUserDeviceType"`
		// 			DeployToMonitor       string    `json:"DeployToMonitor"`
		// 			DeviceListAllowed     string    `json:"DeviceListAllowed"`
		// 			ModelCreatedVersion   string    `json:"ModelCreatedVersion"`
		// 			ModelUpdatedVersion   string    `json:"ModelUpdatedVersion"`
		// 			NameUseredit          string    `json:"name_useredit"`
		// 			OriginalName          string    `json:"OriginalName"`
		// 			PeerNames             []struct {
		// 				Name                        string  `json:"Name"`
		// 				UserDeviceType              string  `json:"UserDeviceType"`
		// 				Percent                     float64 `json:"Percent"`
		// 				Icon                        string  `json:"Icon"`
		// 				UserDeviceTypeDisplayString string  `json:"UserDeviceTypeDisplayString"`
		// 			} `json:"PeerNames"`
		// 			Pending                     string `json:"Pending"`
		// 			Revoked                     string `json:"Revoked"`
		// 			TimelineAllowed             string `json:"TimelineAllowed"`
		// 			TimelineDefault             string `json:"TimelineDefault"`
		// 			Type                        string `json:"Type"`
		// 			UserDeletable               string `json:"UserDeletable"`
		// 			UserDeviceTypeDisplayString string `json:"UserDeviceTypeDisplayString"`
		// 			UserEditable                string `json:"UserEditable"`
		// 			UserEditableMeta            string `json:"UserEditableMeta"`
		// 			UserMergeable               string `json:"UserMergeable"`
		// 			UserShowBubble              string `json:"UserShowBubble"`
		// 			UserShowInDeviceList        string `json:"UserShowInDeviceList"`
		// 		} `json:"tags,omitempty"`
		// 		Tags3 struct {
		// 			Alertable             string    `json:"Alertable"`
		// 			AlwaysOn              string    `json:"AlwaysOn"`
		// 			DateCreated           time.Time `json:"DateCreated"`
		// 			DateFirstUsage        string    `json:"DateFirstUsage"`
		// 			DefaultUserDeviceType string    `json:"DefaultUserDeviceType"`
		// 			DeployToMonitor       string    `json:"DeployToMonitor"`
		// 			DeviceListAllowed     string    `json:"DeviceListAllowed"`
		// 			ModelCreatedVersion   string    `json:"ModelCreatedVersion"`
		// 			ModelUpdatedVersion   string    `json:"ModelUpdatedVersion"`
		// 			NameUseredit          string    `json:"name_useredit"`
		// 			NameUserGuess         string    `json:"NameUserGuess"`
		// 			OriginalName          string    `json:"OriginalName"`
		// 			PeerNames             []struct {
		// 				Name                        string  `json:"Name"`
		// 				UserDeviceType              string  `json:"UserDeviceType"`
		// 				Percent                     float64 `json:"Percent"`
		// 				Icon                        string  `json:"Icon"`
		// 				UserDeviceTypeDisplayString string  `json:"UserDeviceTypeDisplayString"`
		// 			} `json:"PeerNames"`
		// 			Pending                     string `json:"Pending"`
		// 			PreselectionIndex           int    `json:"PreselectionIndex"`
		// 			Revoked                     string `json:"Revoked"`
		// 			TimelineAllowed             string `json:"TimelineAllowed"`
		// 			TimelineDefault             string `json:"TimelineDefault"`
		// 			Type                        string `json:"Type"`
		// 			UserDeletable               string `json:"UserDeletable"`
		// 			UserDeviceType              string `json:"UserDeviceType"`
		// 			UserDeviceTypeDisplayString string `json:"UserDeviceTypeDisplayString"`
		// 			UserEditable                string `json:"UserEditable"`
		// 			UserEditableMeta            string `json:"UserEditableMeta"`
		// 			UserMergeable               string `json:"UserMergeable"`
		// 			UserShowBubble              string `json:"UserShowBubble"`
		// 			UserShowInDeviceList        string `json:"UserShowInDeviceList"`
		// 		} `json:"tags,omitempty"`
		// 		Tags4 struct {
		// 			Alertable             string    `json:"Alertable"`
		// 			AlwaysOn              string    `json:"AlwaysOn"`
		// 			DateCreated           time.Time `json:"DateCreated"`
		// 			DateFirstUsage        string    `json:"DateFirstUsage"`
		// 			DefaultUserDeviceType string    `json:"DefaultUserDeviceType"`
		// 			DeployToMonitor       string    `json:"DeployToMonitor"`
		// 			DeviceListAllowed     string    `json:"DeviceListAllowed"`
		// 			ModelCreatedVersion   string    `json:"ModelCreatedVersion"`
		// 			ModelUpdatedVersion   string    `json:"ModelUpdatedVersion"`
		// 			NameUseredit          string    `json:"name_useredit"`
		// 			NameUserGuess         string    `json:"NameUserGuess"`
		// 			OriginalName          string    `json:"OriginalName"`
		// 			PeerNames             []struct {
		// 				Name                        string  `json:"Name"`
		// 				UserDeviceType              string  `json:"UserDeviceType"`
		// 				Percent                     float64 `json:"Percent"`
		// 				Icon                        string  `json:"Icon"`
		// 				UserDeviceTypeDisplayString string  `json:"UserDeviceTypeDisplayString"`
		// 			} `json:"PeerNames"`
		// 			Pending                     string `json:"Pending"`
		// 			Revoked                     string `json:"Revoked"`
		// 			TimelineAllowed             string `json:"TimelineAllowed"`
		// 			TimelineDefault             string `json:"TimelineDefault"`
		// 			Type                        string `json:"Type"`
		// 			UserDeletable               string `json:"UserDeletable"`
		// 			UserDeviceType              string `json:"UserDeviceType"`
		// 			UserDeviceTypeDisplayString string `json:"UserDeviceTypeDisplayString"`
		// 			UserEditable                string `json:"UserEditable"`
		// 			UserEditableMeta            string `json:"UserEditableMeta"`
		// 			UserMergeable               string `json:"UserMergeable"`
		// 			UserShowBubble              string `json:"UserShowBubble"`
		// 			UserShowInDeviceList        string `json:"UserShowInDeviceList"`
		// 		} `json:"tags,omitempty"`
		// 		Tags5 struct {
		// 			Alertable             string    `json:"Alertable"`
		// 			AlwaysOn              string    `json:"AlwaysOn"`
		// 			DateCreated           time.Time `json:"DateCreated"`
		// 			DateFirstUsage        string    `json:"DateFirstUsage"`
		// 			DefaultUserDeviceType string    `json:"DefaultUserDeviceType"`
		// 			DeployToMonitor       string    `json:"DeployToMonitor"`
		// 			DeviceListAllowed     string    `json:"DeviceListAllowed"`
		// 			ModelCreatedVersion   string    `json:"ModelCreatedVersion"`
		// 			ModelUpdatedVersion   string    `json:"ModelUpdatedVersion"`
		// 			NameUseredit          string    `json:"name_useredit"`
		// 			OriginalName          string    `json:"OriginalName"`
		// 			PeerNames             []struct {
		// 				Name                        string  `json:"Name"`
		// 				UserDeviceType              string  `json:"UserDeviceType"`
		// 				Percent                     float64 `json:"Percent"`
		// 				Icon                        string  `json:"Icon"`
		// 				UserDeviceTypeDisplayString string  `json:"UserDeviceTypeDisplayString"`
		// 			} `json:"PeerNames"`
		// 			Pending                     string `json:"Pending"`
		// 			Revoked                     string `json:"Revoked"`
		// 			TimelineAllowed             string `json:"TimelineAllowed"`
		// 			TimelineDefault             string `json:"TimelineDefault"`
		// 			Type                        string `json:"Type"`
		// 			UserDeletable               string `json:"UserDeletable"`
		// 			UserDeviceTypeDisplayString string `json:"UserDeviceTypeDisplayString"`
		// 			UserEditable                string `json:"UserEditable"`
		// 			UserEditableMeta            string `json:"UserEditableMeta"`
		// 			UserMergeable               string `json:"UserMergeable"`
		// 			UserShowBubble              string `json:"UserShowBubble"`
		// 			UserShowInDeviceList        string `json:"UserShowInDeviceList"`
		// 		} `json:"tags,omitempty"`
		// 		Make  string `json:"make,omitempty"`
		// 		Tags6 struct {
		// 			Alertable                   string        `json:"Alertable"`
		// 			AlwaysOn                    string        `json:"AlwaysOn"`
		// 			DateCreated                 time.Time     `json:"DateCreated"`
		// 			DateFirstUsage              string        `json:"DateFirstUsage"`
		// 			DefaultUserDeviceType       string        `json:"DefaultUserDeviceType"`
		// 			DeployToMonitor             string        `json:"DeployToMonitor"`
		// 			DeviceListAllowed           string        `json:"DeviceListAllowed"`
		// 			MergedDevices               string        `json:"MergedDevices"`
		// 			ModelCreatedVersion         string        `json:"ModelCreatedVersion"`
		// 			ModelUpdatedVersion         string        `json:"ModelUpdatedVersion"`
		// 			NameUseredit                string        `json:"name_useredit"`
		// 			OriginalName                string        `json:"OriginalName"`
		// 			PeerNames                   []interface{} `json:"PeerNames"`
		// 			Pending                     string        `json:"Pending"`
		// 			Revoked                     string        `json:"Revoked"`
		// 			TimelineAllowed             string        `json:"TimelineAllowed"`
		// 			TimelineDefault             string        `json:"TimelineDefault"`
		// 			Type                        string        `json:"Type"`
		// 			UserDeletable               string        `json:"UserDeletable"`
		// 			UserDeviceType              string        `json:"UserDeviceType"`
		// 			UserDeviceTypeDisplayString string        `json:"UserDeviceTypeDisplayString"`
		// 			UserEditable                string        `json:"UserEditable"`
		// 			UserEditableMeta            string        `json:"UserEditableMeta"`
		// 			UserMergeable               string        `json:"UserMergeable"`
		// 			UserShowBubble              string        `json:"UserShowBubble"`
		// 			UserShowInDeviceList        string        `json:"UserShowInDeviceList"`
		// 			Virtual                     string        `json:"Virtual"`
		// 		} `json:"tags,omitempty"`
		// 		GivenMake string `json:"given_make,omitempty"`
		// 		MonitorID int    `json:"monitorId,omitempty"`
		// 		Tags7     struct {
		// 			UserDeviceTypeDisplayString string `json:"UserDeviceTypeDisplayString"`
		// 		} `json:"tags,omitempty"`
		// 	} `json:"devices"`
		TotalCost int `json:"total_cost"`
		// 	TotalCosts []int `json:"total_costs"`
	} `json:"consumption"`
	Production struct {
		Total float64 `json:"total"`
		// 	Totals  []float64 `json:"totals"`
		// 	Devices []struct {
		// 		ID   string `json:"id"`
		// 		Name string `json:"name"`
		// 		Icon string `json:"icon"`
		// 		Tags struct {
		// 			DefaultUserDeviceType       string `json:"DefaultUserDeviceType"`
		// 			DeviceListAllowed           string `json:"DeviceListAllowed"`
		// 			TimelineAllowed             string `json:"TimelineAllowed"`
		// 			UserDeleted                 string `json:"UserDeleted"`
		// 			UserDeviceTypeDisplayString string `json:"UserDeviceTypeDisplayString"`
		// 			UserEditable                string `json:"UserEditable"`
		// 			UserMergeable               string `json:"UserMergeable"`
		// 			UserShowBubble              string `json:"UserShowBubble"`
		// 			UserShowInDeviceList        string `json:"UserShowInDeviceList"`
		// 		} `json:"tags"`
		// 		History     []float64 `json:"history"`
		// 		Avgw        float64   `json:"avgw"`
		// 		TotalCost   int       `json:"total_cost"`
		// 		Pct         float64   `json:"pct"`
		// 		CostHistory []int     `json:"cost_history"`
		// 	} `json:"devices"`
		TotalCost int `json:"total_cost"`
		// 	TotalCosts []int `json:"total_costs"`
	} `json:"production"`
	// GridToBattery              interface{} `json:"grid_to_battery"`
	// SolarToHome                interface{} `json:"solar_to_home"`
	// SolarToBattery             interface{} `json:"solar_to_battery"`
	// BatteryToHome              interface{} `json:"battery_to_home"`
	// BatteryToGrid              interface{} `json:"battery_to_grid"`
	// TopMovers                  interface{} `json:"top_movers"`
	ToGrid   float64 `json:"to_grid"`
	FromGrid float64 `json:"from_grid"`
	// ConsumptionCostChangeCents interface{} `json:"consumption_cost_change_cents"`
	// ConsumptionPercentChange   interface{} `json:"consumption_percent_change"`
	// ProductionPercentChange    interface{} `json:"production_percent_change"`
	ToGridCost   int `json:"to_grid_cost"`
	FromGridCost int `json:"from_grid_cost"`
	// TrendText                  interface{} `json:"trend_text"`
	// UsageText                  interface{} `json:"usage_text"`
	// TrendConsumption           interface{} `json:"trend_consumption"`
	// TrendCost                  interface{} `json:"trend_cost"`
	Scale         string  `json:"scale"`
	SolarPowered  int     `json:"solar_powered"`
	NetProduction float64 `json:"net_production"`
	ProductionPct int     `json:"production_pct"`
	// ConsumptionKwhChange       interface{} `json:"consumption_kwh_change"`
}

type SenseAuth struct {
	Authorized  bool   `json:"authorized"`
	AccountID   int    `json:"account_id"`
	UserID      int    `json:"user_id"`
	AccessToken string `json:"access_token"`
}

func authSense() string {
	authURL := "https://api.sense.com/apiservice/api/v1/authenticate"
	method := "POST"

	payload := strings.NewReader("email=" + url.QueryEscape(config.String("sense.username")) + "&password=" + url.QueryEscape(config.String("sense.password")))

	client := &http.Client{}
	req, err := http.NewRequest(method, authURL, payload)

	if err != nil {
		splunkLogger.Infoln(err)
		return ""
	}
	req.Header.Add("Sense-Client-Version", "1.17.1-20c25f9")
	req.Header.Add("X-Sense-Protocol", "3")
	req.Header.Add("User-Agent", "okhttp/3.8.0")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		splunkLogger.WithFields(log.Fields{"Error": "Error making HTTP call to Sense API"}).Error(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		splunkLogger.WithFields(log.Fields{"Error": "Error getting response body in HTTP call to Sense API"}).Error(err)
		return ""
	}

	authResponse := SenseAuth{}

	err = json.Unmarshal(body, &authResponse)
	if err != nil || !(authResponse.Authorized) {
		splunkLogger.WithFields(log.Fields{"Error": "Error getting response body in HTTP call to Sense API"}).Error(err)
		return ""
	}
	return authResponse.AccessToken
}

func writeToInfluxDB(c influxclient.Client, pointName string, tags map[string]string,
	fields map[string]interface{}, t time.Time) {

	bp, nBPError := influxclient.NewBatchPoints(influxclient.BatchPointsConfig{
		Database: config.String("influxdb.db"),
	})

	if nBPError != nil {
		splunkLogger.WithFields(log.Fields{"Error": "Error creating Batchpoints with config"}).Error(nBPError)

	}

	// 	fmt.Println(bp)

	p, _ := influxclient.NewPoint(pointName, tags, fields, t)

	bp.AddPoint(p)

	writeErr := c.Write(bp)
	if writeErr != nil {
		splunkLogger.WithFields(log.Fields{"Error": "Expected to see no error and didn't"}).Error(writeErr)

	} else {
		// log.Debugf("Wrote %v into InfluxDB with tags %v and value: %v at %v\n", pointName, tags, fields, t)
		splunkLogger.WithFields(fields).WithField("point", pointName).WithField("fields", fmt.Sprint(fields)).WithField("tags", fmt.Sprint(tags)).Debug("Wrote point into InfluxDB")

	}

}

func setupConfig() {

	config.WithOptions(config.ParseEnv)
	config.AddDriver(yaml.Driver)
	err := config.LoadFiles(configFilePath)
	if err != nil {
		splunkLogger.WithFields(log.Fields{"Error": "Error loading config file"}).Error(err)

	}

}

// 6 months token: https://enlighten.enphaseenergy.com/entrez-auth-token?serial_num=SERIALNUMBER
func getLongLivedJWT() (JWTToken, error) {

	splunkLogger.Info("Getting long lived JWT token")

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	// First, login using your username and password
	fieldsLogin := url.Values{"user[email]": {config.String("enphase.EnphaseUser")}, "user[password]": {config.String("enphase.EnphasePassword")}}

	_, errLogin := client.PostForm("https://enlighten.enphaseenergy.com//login/login", fieldsLogin)

	if errLogin != nil {
		splunkLogger.WithField("Error", errLogin).Fatalf("Error loggin in to get long term JWT")

	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://enlighten.enphaseenergy.com/entrez-auth-token?serial_num=%s", config.String("enphase.EnphaseEnvoySerial")), nil)
	requestResponse, requestError := client.Do(req)
	if requestError != nil {
		splunkLogger.WithFields(log.Fields{"Error": "Issue loging-in getting long term JWT", "Enphase Serial": config.String("enphase.EnphaseEnvoySerial")}).Error(requestError)
		panic(0)

	}
	splunkLogger.WithField("response", requestResponse).WithField("EnphaseSerial", config.String("enphase.EnphaseEnvoySerial")).Debugln("Response from Enphase entrez-auth-token")

	defer requestResponse.Body.Close()

	body, err := ioutil.ReadAll(requestResponse.Body)
	if err != nil {
		splunkLogger.WithField("Error", err).Fatalln("Error reading response body:" + requestResponse.Status)

	}

	jwtToken := JWTToken{}
	unmarshalError := json.Unmarshal([]byte(body), &jwtToken)

	if unmarshalError != nil {
		splunkLogger.WithFields(log.Fields{"responseBody": string(body), "unmarshalError": unmarshalError}).Fatalln("Error unmarshalling Sense data")
	}
	splunkLogger.Infoln("Retrieved long-lived JWT token successfully")
	return jwtToken, nil

}

// The EPScraper struct just has some config data in it, it should be pretty
// evident from the variable names what the data is
func getJWT() (string, error) {
	// Error handling stripped for brevity

	// We need to have an HTTP client that stores cookies, for reasons that
	// will become apparent
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	// First, login using your username and password
	fieldsLogin := url.Values{"username": {config.String("enphase.EnphaseUser")}, "password": {config.String("enphase.EnphasePassword")}}

	_, errLogin := client.PostForm("https://entrez.enphaseenergy.com/login", fieldsLogin)

	if errLogin != nil {
		splunkLogger.WithField("Error", errLogin).Fatalf("Error loggin in to Enphase")
	}

	// Second give system parameters
	resp2, _ := client.PostForm("https://entrez.enphaseenergy.com/entrez_tokens",
		url.Values{"Site": {config.String("enphase.EnphaseSite")}, "serialNum": {config.String("enphase.EnphaseEnvoySerial")}})
	// htmlquery is like an xpath library
	doc, _ := htmlquery.Parse(resp2.Body)
	textareas := htmlquery.Find(doc, "//textarea[@id=\"JWTToken\"]")
	// The first and only textarea with the JWTToken id is the JWT
	splunkLogger.Infoln("Retrieved short-lived JWT token successfully")

	return htmlquery.InnerText(textareas[0]), nil
}

var influxDBcnx influxclient.Client

func connectToInfluxDB() (influxclient.Client, error) {

	c, err := influxclient.NewHTTPClient(influxclient.HTTPConfig{
		Addr:     config.String("influxdb.host"),
		Username: config.String("influxdb.user"),
		Password: config.String("influxdb.password"),
	})
	if err != nil {
		log.Fatalf("Error creating InfluxDB Client: %s", err.Error())
	}

	return c, err
}

func initInfluxDB() influxclient.Client {
	influxDBcnx, influxdbcnxerror := connectToInfluxDB()

	if influxdbcnxerror != nil {
		splunkLogger.WithField("Error", influxdbcnxerror).Infoln("Couldn't connect to InfluxDB")
	}

	splunkLogger.Infoln("Connected successfully to influxdb")

	return influxDBcnx
}

func scheduleInserts(myJWTtoken string) {

	period := time.Duration(config.Int("influxdb.periodInMinutes"))
	ticker := time.NewTicker(period * time.Minute)

	senseToken := authSense()

	loadEnphaseDataAndWriteItToInfluxDB(myJWTtoken)

	if senseToken == "" {
		splunkLogger.Infoln("No Sense token, not getting Sense data")
		return
	}

	if config.Bool("sense.enabled") && senseToken != "" {
		senseData := loadSenseData(senseToken)
		if senseData != nil {
			writeSenseDataToInfluxDB(*senseData)
		}
	}

	for range ticker.C {

		loadEnphaseDataAndWriteItToInfluxDB(myJWTtoken)

		if config.Bool("sense.enabled") && senseToken != "" {
			senseData := loadSenseData(senseToken)
			if senseData != nil {
				writeSenseDataToInfluxDB(*senseData)
			}

		}
	}
}
func writeSenseDataToInfluxDB(senseTrendsData SenseTrends) {
	splunkLogger.Infoln("Writing Sense data to InfluxDB")
	eventTime := time.Now()

	tags := map[string]string{"senseMonitorID": config.String("sense.monitorID")}

	fields := map[string]interface{}{
		"Production":    senseTrendsData.Production.Total,
		"Consumption":   senseTrendsData.Consumption.Total,
		"ToGrid":        senseTrendsData.ToGrid,
		"FromGrid":      senseTrendsData.FromGrid,
		"SolarPowered":  senseTrendsData.SolarPowered,
		"NetProduction": senseTrendsData.NetProduction,
		"ProductionPct": senseTrendsData.ProductionPct,
	}
	writeToInfluxDB(influxDBcnx, "sense", tags, fields, eventTime)

}

func loadSenseData(senseToken string) *SenseTrends {

	beginingOfDay := time.Now().Round(24 * time.Hour)

	url := "https://api.sense.com/apiservice/api/v1/app/history/trends?monitor_id=" + config.String("sense.monitorID") + "&scale=DAY&start=" + beginingOfDay.Format("2006-01-02T15:04:05.000Z")
	splunkLogger.WithFields(log.Fields{"URL": url}).Infoln("Retrieving Sense data from unofficial API")

	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		splunkLogger.WithField("Error", err).WithField("monitor_id", config.String("sense.monitorID")).WithField("URL", url).Error("Error retrieving tren sense data")
		return nil
	}
	req.Header.Add("Authorization", "Bearer "+senseToken)

	res, err := client.Do(req)
	if err != nil {
		splunkLogger.Error(err)
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		splunkLogger.WithFields(log.Fields{"responseStatusCode": res.StatusCode}).Infoln("Got a non-200 response from Sense API")

		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		splunkLogger.Error(err)
		return nil
	}

	senseTrendsData := SenseTrends{}
	unmarshalError := json.Unmarshal([]byte(body), &senseTrendsData)

	if unmarshalError != nil {

		splunkLogger.WithFields(log.Fields{"responseBody": string(body), "unmarshalError": unmarshalError}).Fatalln("Error unmarshalling Sense data")

		// panic(0)
	}
	splunkLogger.WithFields(log.Fields{
		"Production": senseTrendsData.Production.Total, "Consumption": senseTrendsData.Consumption.Total, "ToGrid": senseTrendsData.ToGrid, "FromGrid": senseTrendsData.FromGrid, "SolarPowered": senseTrendsData.SolarPowered, "NetProduction": senseTrendsData.NetProduction, "ProductionPct": senseTrendsData.ProductionPct}).Infoln("Retrieved Sense Trends data successfully")
	return &senseTrendsData

}

func loadEnphaseDataAndWriteItToInfluxDB(myJWTtoken string) {
	log.Infoln("Retrieving Enphase Production data, from local endpoint")
	enphaseData := loadProductionDetailsData(myJWTtoken)

	for _, data := range enphaseData.Production {

		eventTime := time.Now()

		tags := map[string]string{"serial": config.String("enphase.EnphaseEnvoySerial"), "type": data.Type}

		fields := map[string]interface{}{
			"whLifetime":           int(data.WhLifetime),
			"WhLastSevenDays":      data.WhLastSevenDays,
			"WhToday":              data.WhToday,
			"WNow":                 int(data.WNow),
			"activeInverterCounts": data.ActiveCount,
		}

		writeToInfluxDB(influxDBcnx, "production", tags, fields, eventTime)

		// b, _ := json.Marshal(fields)
		// fmt.Println(string(b))

		splunkLogger.WithFields(fields).Infoln("Wrote Enphase Production data to InfluxDB")

		// log.Infof("Today's Production (%s): %fWh", data.Type, data.WhToday)
		// log.Infof("Week's Production (%s): %fWh", data.Type, data.WhLastSevenDays)
		// log.Infof("Lifetime Production (%s): %dWh", data.Type, data.WhLifetime)
		// log.Infof("Current Production (%s): %dW", data.Type, data.WNow)
		// log.Infof("Active Inverters (%s): %d", data.Type, data.ActiveCount)

	}

	for _, data := range enphaseData.Consumption {

		eventTime := time.Now()

		tags := map[string]string{"serial": config.String("enphase.EnphaseEnvoySerial"), "type": data.MeasurementType}

		fields := map[string]interface{}{
			"whLifetime":      data.WhLifetime,
			"WhLastSevenDays": data.WhLastSevenDays,
			"WhToday":         data.WhToday,
		}
		splunkLogger.Debugf("WhLifeTime: \n %#v\n", data.WhLifetime)
		// log.Debug(tags, fields)

		writeToInfluxDB(influxDBcnx, "consumption", tags, fields, eventTime)

		// log.Infof("Today's Consumption (%s): %f", data.MeasurementType, data.WhToday)
		splunkLogger.WithFields(log.Fields{"MeasurementType": data.MeasurementType, "WhToday": data.WhToday}).Debugln("Today's Consumption")
		// log.Infof("Week's Consumption (%s): %f", data.MeasurementType, data.WhLastSevenDays)
		splunkLogger.WithFields(log.Fields{"MeasurementType": data.MeasurementType, "WhLastSevenDays": data.WhLastSevenDays}).Debugln("Week's Consumption")
		// log.Infof("Lifetime Consumption (%s): %f", data.MeasurementType, data.WhLifetime)
		splunkLogger.WithFields(log.Fields{"MeasurementType": data.MeasurementType, "WhLastSevenDays": data.WhLifetime}).Debugln("Lifetime Consumption")

	}

	splunkLogger.Infoln("Retrieving Enphase Inverter data, from local endpoint")
	invertersData := loadInverterData(myJWTtoken)

	totalInverters := 0
	for _, data_inverter := range invertersData {

		// eventTime := time.Now()
		eventTime := time.Unix(int64(data_inverter.Lastreportdate), 0)

		inverterSerial := data_inverter.Serialnumber

		inverterID := inverterSerial[len(inverterSerial)-5:]

		tags := map[string]string{"serial": config.String("enphase.EnphaseEnvoySerial"), "inverter": inverterID}

		fields := map[string]interface{}{
			"lastReportWatts": data_inverter.Lastreportwatts,
			"maxReportWatts":  data_inverter.Maxreportwatts,
		}

		splunkLogger.WithFields(fields).WithField("inverter", inverterID).WithField("tags", fmt.Sprint(tags)).Debug("Inverter data")

		writeToInfluxDB(influxDBcnx, "inverters", tags, fields, eventTime)
		totalInverters += data_inverter.Lastreportwatts

	}

	// if !debug {
	splunkLogger.WithField("TotalReportedWatts", totalInverters).Debug("Total Reported Watts for Inverters")
	// }

}

func writeConfig() error {

	splunkLogger.Infoln("Writing config to file")

	buf := new(bytes.Buffer)

	_, dumpError := config.DumpTo(buf, config.Yaml)
	if dumpError != nil {
		return dumpError
	}
	writeError := ioutil.WriteFile(configFilePath, buf.Bytes(), 0755)
	if writeError != nil {
		return writeError
	}
	return nil

}

func loadJWTIntoCookie(jwt string) (cookiejar.Jar, error) {
	// You need to empty the cookie jar before trying to load a new JWT in.
	// If you have an old cookie from a previous JWT auth, it gets confused and you don't end
	// up "refreshing" your auth. Then your system breaks when you hit the expiry time
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalln(err)
	}
	// Erase the cookies in the scraper's client bc the envoy doesn't seem to overwrite the
	// jwt if there is one in the session already
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr,
		Jar: jar,
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/auth/check_jwt", config.String("enphase.EnvoyHost")), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	requestResponse, requestError := client.Do(req)
	if requestError != nil {
		splunkLogger.Fatalln(requestError)
	}
	splunkLogger.WithField("Response", requestResponse).Debug("Making call to auth/check_jwt")
	// again, all error handling stripped. You'd normally check that response

	// All we needed was the cookie, which is now in ep.client.Jar
	splunkLogger.Infoln("Retrieved cookies successfully from JWT auth page")

	return *jar, nil
}
func loadProductionDetailsData(myJWTtoken string) enphaseMetrics {
	// /production.json?details=1

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/production.json?details=1", config.String("enphase.EnvoyHost")), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", myJWTtoken))

	requestResponse, requestError := client.Do(req)
	if requestError != nil {
		log.Fatalln(requestError)
	}

	defer requestResponse.Body.Close()

	body, err := ioutil.ReadAll(requestResponse.Body)
	if err != nil {
		splunkLogger.Fatalln(err)
	}

	splunkLogger.WithFields(log.Fields{"Response": body}).Debug("Production details response from Enphase API")

	res := enphaseMetrics{}
	unMarshalError := json.Unmarshal([]byte(body), &res)

	if unMarshalError != nil {

		splunkLogger.WithFields(log.Fields{"responseBody": string(body), "unmarshalError": unMarshalError}).Fatalln("Error unmarshalling Sense data")
	}

	splunkLogger.Infoln("Retrieved Enphase production data successfully from local endpoint")

	return res

}

func loadInverterData(myJWTtoken string) Inverters {
	// /api/v1/production/inverters
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/production/inverters", config.String("enphase.EnvoyHost")), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", myJWTtoken))

	requestResponse, requestError := client.Do(req)
	if requestError != nil {
		log.Fatalln(requestError)
	}
	log.Debugln(requestResponse)

	defer requestResponse.Body.Close()

	body, err := ioutil.ReadAll(requestResponse.Body)
	if err != nil {
		log.Fatalln(err)
	}

	res := Inverters{}
	unmarshalError := json.Unmarshal([]byte(body), &res)

	if unmarshalError != nil {
		splunkLogger.WithFields(log.Fields{"responseBody": string(body), "unmarshalError": unmarshalError}).Fatalln("Error unmarshalling Sense data")
	}

	log.Infoln("Retrieved Enphase inverter data successfully from local endpoint")

	return res

}

func loadStreamData(authedCookieJar cookiejar.Jar) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr,
		Jar: &authedCookieJar,
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/stream/meter", config.String("enphase.EnvoyHost")), nil)
	// req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.String("enphase.jwtToken.Token")))
	// req.SetBasicAuth("installer", "107334")

	requestResponse, requestError := client.Do(req)
	if requestError != nil {
		log.Fatalln(requestError)
	}
	log.Debugln(requestResponse)

	defer requestResponse.Body.Close()

	body, err := ioutil.ReadAll(requestResponse.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Debugln(body)

	log.Infoln("Requested Stream data from local endpoint, Response status was: ", requestResponse.Status)

}

var splunkLogger *log.Entry

func initLoggers() {

	const APPNAME = "enphaselocal2influx"
	if config.String("logformat") == "JSON" {
		log.SetFormatter(&log.JSONFormatter{})
	}
	log.SetOutput(os.Stdout)

	log.SetLevel(log.Level(config.Int("loglevel")))

	// if config.Int("loglevel") > 4 {
	log.SetReportCaller(true)
	// }

	splunkLogger = log.WithFields(log.Fields{"app": APPNAME})

	splunkLogger.Info("Loggers initialized")

}

func main() {

	setupConfig()
	initLoggers()

	splunkLogger.WithField("config", fmt.Sprint(config.Data())).Debug("Config loaded")

	splunkLogger.Debug("Config loaded")

	tokenExpiry, intError := strconv.Atoi(config.String("enphase.jwtToken.ExpiresAt"))
	tokenGen, intError2 := strconv.Atoi(config.String("enphase.jwtToken.GenerationTime"))
	token := config.String("enphase.jwtToken.Token")

	var longLivedJWT JWTToken

	if token == "" || intError != nil || intError2 != nil || time.Unix(int64(tokenExpiry), 0).Before(time.Now()) {
		splunkLogger.Infoln("No JWT token found in config")
		longLivedJWT, _ = getLongLivedJWT()
		splunkLogger.Debugf("Long lived JWT: \n %#v\n", longLivedJWT)

		// Not sure if I love this
		config.Set("enphase.jwtToken.Token", longLivedJWT.Token)
		config.Set("enphase.jwtToken.ExpiresAt", longLivedJWT.ExpiresAt)
		config.Set("enphase.jwtToken.GenerationTime", longLivedJWT.GenerationTime)

		writeConfigError := writeConfig()
		if writeConfigError != nil {
			splunkLogger.WithFields(log.Fields{"writeConfigError": writeConfigError}).Fatalln("Error writing config file")
		}

	} else {
		longLivedJWT = JWTToken{token, tokenExpiry, tokenGen}

		splunkLogger.WithFields(log.Fields{"JWT_Expiration": time.Unix(int64(tokenExpiry), 0).String()}).Infoln("Using stored JWT token")

	}

	influxDBcnx = initInfluxDB()

	scheduleInserts(longLivedJWT.Token)

}
