package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	gheHost        string        = "my.github-enterprise.com"
	gheOrg         string        = "MyOrg"
	appId          int           = 123   // id of github application
	appInstallId   int           = 12345 // id of github app installation
	appPrivateKey  string        = "/path/to/private-key.pem"
	jwtTtl         time.Duration = time.Second * 300
	jwtIssueBuffer time.Duration = time.Second * 30
)

type AppAccessResp struct {
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expires_at"`
	Permissions struct {
		OrganizationSelfHostedRunners string `json:"organization_self_hosted_runners"`
	} `json:"permissions"`
	RepositorySelection string `json:"repository_selection"`
}

type RunnerRegResp struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func main() {
	// log warning
	log.Println("WARNING: !!! NOT FOR USE IN PROD !!! This is considered a proof-of-concept only! Secrets are being logged!")

	// read private key file
	privateKey, err := ioutil.ReadFile(appPrivateKey)
	if err != nil {
		log.Fatalln(err)
	}

	// create new jwt token
	// https://docs.github.com/en/enterprise-server@3.4/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-a-github-app
	jwtToken, err := CreateJWT(privateKey, appId)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("JWT TOKEN:", jwtToken)

	// get app installation access token
	// https://docs.github.com/en/enterprise-server@3.4/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-an-installation
	appToken, err := GetAppAccessToken(jwtToken, appInstallId)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("APP TOKEN:", appToken)

	// get runner registration token
	// https://docs.github.com/en/enterprise-server@3.4/rest/actions/self-hosted-runners#create-a-registration-token-for-an-organization
	runnerRegToken, err := GetRunnerRegToken(appToken, gheOrg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("RUNNER REG TOKEN:", runnerRegToken)
}

func SetupHttpReq(url string, method string, tokenType string, token string) (*retryablehttp.Client, *retryablehttp.Request) {
	// setup client
	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.RetryWaitMin = 30 * time.Second
	client.Logger = nil

	// setup request
	req, err := retryablehttp.NewRequest(method, url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	// setup headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", tokenType, token))

	return client, req
}

func DoHttpReq(client *retryablehttp.Client, req *retryablehttp.Request) ([]byte, error) {
	// do the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)

	// get response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// ensure response code is what we expect
	if resp.StatusCode != 201 {
		log.Fatalln("something broke", resp.StatusCode, string(body))
	}

	return body, nil
}

func CreateJWT(privateKey []byte, issuer int) (string, error) {
	// parse RSA private key
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return "", fmt.Errorf("create: parse key: %w", err)
	}

	// setup claims
	now := time.Now().UTC()
	claims := make(jwt.MapClaims)
	claims["iss"] = issuer
	claims["exp"] = now.Add(jwtTtl).Unix()
	claims["iat"] = now.Add(-jwtIssueBuffer).Unix()
	claims["nbf"] = now.Add(-jwtIssueBuffer).Unix()

	// create jwt token
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		return "", err
	}

	return token, nil
}

func GetAppAccessToken(jwtToken string, appInstall int) (string, error) {
	// setup url string
	url := fmt.Sprintf("https://%s/api/v3/app/installations/%d/access_tokens", gheHost, appInstall)

	// create client and req objects
	client, req := SetupHttpReq(url, http.MethodPost, "Bearer", jwtToken)

	// do the request
	body, err := DoHttpReq(client, req)
	if err != nil {
		return "", err
	}

	// unmarshall json response
	var appAccessResp AppAccessResp
	json.Unmarshal(body, &appAccessResp)

	return appAccessResp.Token, nil
}

func GetRunnerRegToken(installToken string, gheOrg string) (string, error) {
	// setup url string
	url := fmt.Sprintf("https://%s/api/v3/orgs/%s/actions/runners/registration-token", gheHost, gheOrg)

	// create client and req objects
	client, req := SetupHttpReq(url, http.MethodPost, "token", installToken)

	// do the request
	body, err := DoHttpReq(client, req)
	if err != nil {
		return "", err
	}

	// unmarshall json response
	var runnerRegResp RunnerRegResp
	json.Unmarshal(body, &runnerRegResp)

	return runnerRegResp.Token, nil
}
