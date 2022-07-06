package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	"github.com/antchfx/htmlquery"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	influxclient "github.com/influxdata/influxdb1-client/v2"
)

const configFilePath = "config.yaml"

var debug = false

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
	WNow             int     `json:"wNow"`
	WhLifetime       int     `json:"whLifetime"`
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

func writeToInfluxDB(c influxclient.Client, pointName string, tags map[string]string,
	fields map[string]interface{}, t time.Time) {

	bp, nBPError := influxclient.NewBatchPoints(influxclient.BatchPointsConfig{
		Database: config.String("influxdb.db"),
	})

	if nBPError != nil {
		log.Println("Error creating Batchpoints with config: ", nBPError)

	}

	// 	fmt.Println(bp)

	p, _ := influxclient.NewPoint(pointName, tags, fields, t)

	bp.AddPoint(p)

	writeErr := c.Write(bp)
	if writeErr != nil {
		log.Printf("unexpected error.  expected %v, actual %v", nil, writeErr)
	} else if debug {
		log.Printf("Wrote %v into InfluxDB with tags %v and value: %v at %v\n", pointName, tags, fields, t)

	}

}

func setupConfig() {

	config.WithOptions(config.ParseEnv)
	config.AddDriver(yaml.Driver)
	err := config.LoadFiles(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	if err != nil {
		log.Fatalln(err)
	}

}

// 6 months token: https://enlighten.enphaseenergy.com/entrez-auth-token?serial_num=SERIALNUMBER
func getLongLivedJWT() (JWTToken, error) {

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	// First, login using your username and password
	fieldsLogin := url.Values{"user[email]": {config.String("enphase.EnphaseUser")}, "user[password]": {config.String("enphase.EnphasePassword")}}

	_, errLogin := client.PostForm("https://enlighten.enphaseenergy.com//login/login", fieldsLogin)
	// Response error checking omitted, but what we needed was the cookie, which is now in the jar
	// fmt.Println(respLogin)
	// fmt.Println(errLogin)
	// fmt.Println(jar)
	if errLogin != nil {
		log.Println("Error loggin in to get long term JWT")
		log.Fatalln(errLogin)
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://enlighten.enphaseenergy.com/entrez-auth-token?serial_num=%s", config.String("enphase.EnphaseEnvoySerial")), nil)
	requestResponse, requestError := client.Do(req)
	if requestError != nil {
		log.Println("Error loging-in getting long term JWT for serial number" + config.String("enphase.EnphaseEnvoySerial"))

		log.Println(requestError)
		panic(0)

	}
	if debug {
		log.Println(requestResponse)
	}

	defer requestResponse.Body.Close()

	body, err := ioutil.ReadAll(requestResponse.Body)
	if err != nil {
		log.Println("Error reading response body:" + requestResponse.Status)

		log.Println(err)
		panic(0)

	}

	jwtToken := JWTToken{}
	unmarshalError := json.Unmarshal([]byte(body), &jwtToken)

	if unmarshalError != nil {
		log.Println("Error unmarshalling:" + string(body))

		log.Println(unmarshalError)
		panic(0)
	}
	log.Println("Retrieved long-lived JWT token successfully")
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
	// Response error checking omitted, but what we needed was the cookie, which is now in the jar
	// fmt.Println(respLogin)
	// fmt.Println(errLogin)
	// fmt.Println(jar)
	if errLogin != nil {
		log.Fatalln(errLogin)
	}

	// Second give system parameters
	resp2, _ := client.PostForm("https://entrez.enphaseenergy.com/entrez_tokens",
		url.Values{"Site": {config.String("enphase.EnphaseSite")}, "serialNum": {config.String("enphase.EnphaseEnvoySerial")}})
	// htmlquery is like an xpath library
	doc, _ := htmlquery.Parse(resp2.Body)
	textareas := htmlquery.Find(doc, "//textarea[@id=\"JWTToken\"]")
	// The first and only textarea with the JWTToken id is the JWT
	log.Println("Retrieved short-lived JWT token successfully")

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
		log.Println("Couldn't connect to InfluxDB")
		log.Fatalln(influxdbcnxerror)
	}

	log.Println("Connected successfully to influxdb")

	return influxDBcnx
}

func scheduleInserts(myCookies cookiejar.Jar) {

	period := time.Duration(config.Int("influxdb.periodInMinutes"))
	ticker := time.NewTicker(period * time.Minute)

	loadDataAndWriteToInfluxDB(myCookies)

	for range ticker.C {
		loadDataAndWriteToInfluxDB(myCookies)
	}
}

func loadDataAndWriteToInfluxDB(myCookies cookiejar.Jar) {
	log.Println("Retrieving Enphase Production data, from local endpoint")
	enphaseData := loadProductionDetailsData(myCookies)

	for _, data := range enphaseData.Production {

		eventTime := time.Now()

		tags := map[string]string{"serial": config.String("enphase.EnphaseEnvoySerial"), "type": data.Type}

		fields := map[string]interface{}{
			"whLifetime":      data.WhLifetime,
			"WhLastSevenDays": data.WhLastSevenDays,
			"WhToday":         data.WhToday,
		}
		if debug {
			log.Printf("WhLifeTime: \n %#v\n", data.WhLifetime)
			log.Print(tags, fields)
		}
		writeToInfluxDB(influxDBcnx, "production", tags, fields, eventTime)

		log.Printf("Today's Production (%s): %f", data.Type, data.WhToday)
		log.Printf("Week's Production (%s): %f", data.Type, data.WhLastSevenDays)
		log.Printf("Lifetime Production (%s): %d", data.Type, data.WhLifetime)

	}

	for _, data := range enphaseData.Consumption {

		eventTime := time.Now()

		tags := map[string]string{"serial": config.String("enphase.EnphaseEnvoySerial"), "type": data.MeasurementType}

		fields := map[string]interface{}{
			"whLifetime":      data.WhLifetime,
			"WhLastSevenDays": data.WhLastSevenDays,
			"WhToday":         data.WhToday,
		}
		if debug {
			log.Printf("WhLifeTime: \n %#v\n", data.WhLifetime)
			log.Print(tags, fields)
		}
		writeToInfluxDB(influxDBcnx, "consumption", tags, fields, eventTime)

		log.Printf("Today's Consumption (%s): %f", data.MeasurementType, data.WhToday)
		log.Printf("Week's Consumption (%s): %f", data.MeasurementType, data.WhLastSevenDays)
		log.Printf("Lifetime Consumption (%s): %f", data.MeasurementType, data.WhLifetime)

	}

	log.Println("Retrieving Enphase Inverter data, from local endpoint")
	invertersData := loadInverterData(myCookies)

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
		if debug {
			log.Printf("Inverter lastReportWatts(%s): \n %#v\n", data_inverter.Serialnumber, data_inverter.Lastreportwatts)
			log.Print(tags, fields)
		}
		writeToInfluxDB(influxDBcnx, "inverters", tags, fields, eventTime)
		totalInverters += data_inverter.Lastreportwatts

	}

	// if !debug {
	log.Printf("Total Reported Watts for Inverters: %d", totalInverters)
	// }

}

func writeConfig() error {

	log.Println("Writing config to file")

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
		log.Fatalln(requestError)
	}
	if debug {
		log.Println(requestResponse)
	}
	// again, all error handling stripped. You'd normally check that response

	// All we needed was the cookie, which is now in ep.client.Jar
	log.Println("Retrieved cookies successfully from JWT auth page")

	return *jar, nil
}
func loadProductionDetailsData(authedCookieJar cookiejar.Jar) enphaseMetrics {
	// /production.json?details=1
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr,
		Jar: &authedCookieJar,
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/production.json?details=1", config.String("enphase.EnvoyHost")), nil)
	// req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	requestResponse, requestError := client.Do(req)
	if requestError != nil {
		log.Fatalln(requestError)
	}
	if debug {
		log.Println(requestResponse)
	}
	defer requestResponse.Body.Close()

	body, err := ioutil.ReadAll(requestResponse.Body)
	if err != nil {
		log.Fatalln(err)
	}

	res := enphaseMetrics{}
	json.Unmarshal([]byte(body), &res)

	log.Println("Retrieved Enphase production data successfully from local endpoint")

	return res

}

func loadInverterData(authedCookieJar cookiejar.Jar) Inverters {
	// /api/v1/production/inverters
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr,
		Jar: &authedCookieJar,
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/production/inverters", config.String("enphase.EnvoyHost")), nil)
	// req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	requestResponse, requestError := client.Do(req)
	if requestError != nil {
		log.Fatalln(requestError)
	}
	if debug {
		log.Println(requestResponse)
	}
	defer requestResponse.Body.Close()

	body, err := ioutil.ReadAll(requestResponse.Body)
	if err != nil {
		log.Fatalln(err)
	}

	res := Inverters{}
	json.Unmarshal([]byte(body), &res)

	log.Println("Retrieved Enphase inverter data successfully from local endpoint")

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
	if debug {
		log.Println(requestResponse)
	}
	defer requestResponse.Body.Close()

	body, err := ioutil.ReadAll(requestResponse.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if debug {
		log.Println(body)
	}

	// res := Inverters{}
	// json.Unmarshal([]byte(body), &res)

	log.Println("Requested Stream data from local endpoint, Response status was: ", requestResponse.Status)
	// log.Println(requestResponse)
	// log.Println(body)

	// return res

}

func main() {

	setupConfig()

	debug = config.Bool("debug")

	if debug {
		log.Printf("config debug: \n %#v\n", debug)
		log.Printf("config data: \n %#v\n", config.Data())
	}

	tokenExpiry, intError := strconv.Atoi(config.String("enphase.jwtToken.ExpiresAt"))
	tokenGen, intError2 := strconv.Atoi(config.String("enphase.jwtToken.GenerationTime"))
	token := config.String("enphase.jwtToken.Token")

	var longLivedJWT JWTToken

	// log.Println(token)
	// log.Println(tokenExpiry)
	// log.Println(tokenGen)
	// log.Println(intError)
	// log.Println(intError2)

	if token == "" || intError != nil || intError2 != nil || time.Unix(int64(tokenExpiry), 0).Before(time.Now()) {
		log.Println("No JWT token found in config")
		longLivedJWT, _ := getLongLivedJWT()
		if debug {
			log.Printf("Long lived JWT: \n %#v\n", longLivedJWT)
		}

		// Not sure if I love this
		config.Set("enphase.jwtToken.Token", longLivedJWT.Token)
		config.Set("enphase.jwtToken.ExpiresAt", longLivedJWT.ExpiresAt)
		config.Set("enphase.jwtToken.GenerationTime", longLivedJWT.GenerationTime)

		writeConfig()

	} else {
		longLivedJWT = JWTToken{token, tokenExpiry, tokenGen}

		log.Println("Using stored JWT token, expires on: " + time.Unix(int64(tokenExpiry), 0).String())

	}

	influxDBcnx = initInfluxDB()

	// value := config.String("enphase.apiKey")

	// config.Set("name", "James")
	// writeConfig()

	// jwt, errorJWT := getJWT()
	// if errorJWT != nil {
	// 	log.Fatalln(errorJWT)
	// }
	// if debug {
	// 	log.Printf("Short Lived JWT: \n %#v\n", jwt)
	// }
	hardToObtainCookies, loadJWTIntoCookieError := loadJWTIntoCookie(longLivedJWT.Token)
	if loadJWTIntoCookieError != nil {
		log.Fatalln(loadJWTIntoCookieError)
	}
	if debug {
		log.Println("Cookies: ")
		log.Println(hardToObtainCookies)
	}

	// name := myConfig.String("name")
	// fmt.Print(name) // "new name"

	// buf := new(bytes.Buffer)

	// _, dumpError := config.DumpTo(buf, config.Yaml)
	// if dumpError != nil {
	// 	log.Fatalln(dumpError)
	// }

	// ioutil.WriteFile("config.yaml", buf.Bytes(), 0755)

	loadStreamData(hardToObtainCookies)

	scheduleInserts(hardToObtainCookies)

}
