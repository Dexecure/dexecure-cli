package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/parnurzeal/gorequest"
	"github.com/tucnak/store"
	"gopkg.in/urfave/cli.v2"
)

var apiEndPoint = "https://dao-api.dexecure.com/api/v1/"
var errorResponse ErrorResponse

func saveToken(token string) {
	var tokenSettings TokenSettings
	tokenSettings.Token = token
	if err := store.Save("token.json", &tokenSettings); err != nil {
		fmt.Println("failed to save the token:", err)
		return
	}
}

func getToken() string {
	var tokenSettings TokenSettings
	store.Load("token.json", &tokenSettings)
	return tokenSettings.Token
}

func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

func parseResponse(body string, res gorequest.Response) Response {

	var response Response

	if res.StatusCode == 200 {
		var responseJSON map[string]interface{}
		json.Unmarshal([]byte(body), &responseJSON)

		// hack 1 for actionhero validation errors
		_, ok := responseJSON["error"].(string)
		if ok {
			response.Error = make(map[string]interface{})
			response.Error["description"] = responseJSON["error"].(string)
			return response
		}

		// hack 2 for actionhero validation errors
		if responseJSON["error"] != nil {
			response.Error = responseJSON["error"].(map[string]interface{})
			return response
		}

		responseStatus := responseJSON["status"].(float64)

		if responseStatus == 200 {
			_, ok := responseJSON["data"].(string)
			if ok {
				response.Data = make(map[string]interface{})
				response.Data["message"] = responseJSON["data"].(string)
			} else {
				response.Data = responseJSON["data"].(map[string]interface{})
			}
		} else {
			response.Error = responseJSON["error"].(map[string]interface{})
		}
	} else {
		fmt.Println("Request to the API failed")
	}

	return response
}

func credentials() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your api token (Visit https://app.dexecure.com/profile?display=api-tokens to generate a token if you don't have one already): ")
	apiTokens, _ := reader.ReadString('\n')

	return strings.TrimSpace(apiTokens)
}

func main() {
	// var token string = ""
	app := &cli.App{}

	app.Name = "Dexecure CLI"
	app.Usage = "Interact with your Dexecure account"
	app.Version = "0.0.2"
	app.Copyright = "Dexecure PTE LTD."
	app.EnableShellCompletion = true

	// config management
	store.Init("dexecure")

	app.Commands = []*cli.Command{
		{
			Name:    "configure",
			Aliases: []string{"c"},
			Usage:   "Add your Dexecure api-tokens",
			Action: func(c *cli.Context) error {
				apiTokens := credentials()
				saveToken(apiTokens)
				fmt.Println("API tokens saved successfully")
				return nil
			},
		},
		{
			Name:    "usage",
			Aliases: []string{"l"},
			Usage:   "Total bandwidth served via Dexecure across all your domains",
			Action: func(c *cli.Context) error {
				if getToken() == "" {
					fmt.Println("API token not found please run \"dexecure-cli configure\"")
					return nil
				}
				res, _, err := gorequest.
					New().
					Get(apiEndPoint+"user").
					Set("Authorization", getToken()).
					End()

				if err != nil {
					fmt.Println(err)
					return nil
				}

				// Read response body
				bdy, er := ioutil.ReadAll(res.Body)
				if er != nil {
					return er
				}
				var user UserResponse
				er = json.Unmarshal(bdy, &user)

				res, _, err = gorequest.
					New().
					Get(apiEndPoint+"user/usage").
					Set("Authorization", getToken()).
					End()

				if err != nil {
					fmt.Println(err)
					return nil
				}

				// Read response body
				bdy, er = ioutil.ReadAll(res.Body)
				if er != nil {
					return er
				}

				var usage UsageResponse
				er = json.Unmarshal(bdy, &usage)

				fmt.Println("Bandwidth Used:")
				fmt.Println(usage.Data.Bandwidth/(1024*1024), "MB of", user.Data.Plan.MaxBandwidth, "GB")
				fmt.Println("Number of Requests:")
				fmt.Println(usage.Data.Requests, "of", user.Data.Plan.MaxRequests)
				fmt.Println("Number of Distributions Used:")
				fmt.Println(usage.Data.Distributions, "of", user.Data.Plan.MaxDistributions)

				return nil
			},
		},
		{
			Name:    "website",
			Aliases: []string{"w"},
			Usage:   "options for managing your website",
			Subcommands: []*cli.Command{
				{
					Name:  "ls",
					Usage: "Get more information about your website",
					Subcommands: []*cli.Command{
						{
							Name:  "id",
							Usage: "information about your website",
							Action: func(c *cli.Context) error {
								if getToken() == "" {
									fmt.Println("API token not found please run \"dexecure-cli configure\"")
									return nil
								}

								id := ""
								if c.Args().Len() > 0 {
									id = c.Args().First()
									id = strings.TrimSpace(id)
									if IsValidUUID(id) == false {
										fmt.Println("Please enter a valid website ID. It must be a valid UUID")
										return nil
									}
								}
								res, _, err := gorequest.
									New().
									Get(fmt.Sprintf("%swebsite/%s", apiEndPoint, id)).
									Set("Authorization", getToken()).
									End()

								if err != nil {
									fmt.Println(err)
									return nil
								}

								// Read response body
								body, er := ioutil.ReadAll(res.Body)
								if er != nil {
									return nil
								}

								// Website Response for /website endpoint
								var wr WebsiteResponse
								// first try to Unmarshal expected response
								er = json.Unmarshal(body, &wr)
								if er != nil {
									// if expected response Unmarshalling failed then
									// try to Unmarshal error
									er = json.Unmarshal(body, &errorResponse)
									if er != nil {
										return nil
									} else {
										fmt.Println("Error: ", errorResponse.Error.Description)
									}
									return nil
								}

								fmt.Println("-----------------------------------------")
								fmt.Println("ID: ", wr.Data.ID)
								fmt.Println("Website URL: ", wr.Data.WebsiteURL)
								fmt.Println("Website Type: ", wr.Data.WebsiteType)
								fmt.Println("Website Name: ", wr.Data.WebsiteName)

								return nil
							},
						},
						{
							Name:  "all",
							Usage: "information about your website(s)",
							Action: func(c *cli.Context) error {
								if getToken() == "" {
									fmt.Println("API token not found please run \"dexecure-cli configure\"")
									return nil
								}
								res, _, err := gorequest.
									New().
									Get(fmt.Sprintf("%swebsite/", apiEndPoint)).
									Set("Authorization", getToken()).
									End()

								if err != nil {
									fmt.Println(err)
									return nil
								}

								// Read response body
								body, er := ioutil.ReadAll(res.Body)
								if er != nil {
									return nil
								}

								// Website Response for /website endpoint
								var wr WebsitesResponse
								// first try to Unmarshal expected response
								er = json.Unmarshal(body, &wr)
								if er != nil {
									// if expected response Unmarshalling failed then
									// try to Unmarshal error
									er = json.Unmarshal(body, &errorResponse)
									if er != nil {
										return nil
									} else {
										fmt.Println("Error: ", errorResponse.Error.Description)
									}
									return nil
								} else {
									fmt.Println("\nTotal number of website:", len(wr.Data))
								}

								if wr.Error.Description != "" {
									fmt.Println(wr.Error.Description)
								} else {
									fmt.Println("")
									for _, website := range wr.Data {
										fmt.Println("-----------------------------------------")
										fmt.Println("ID: ", website.ID)
										fmt.Println("Website URL: ", website.WebsiteURL)
										fmt.Println("Website Type: ", website.WebsiteType)
										fmt.Println("Website Name: ", website.WebsiteName)
									}
									fmt.Println("-----------------------------------------")
								}

								return nil
							},
						},
					},
				},
				{
					Name:  "add",
					Usage: "add a new website",
					Action: func(c *cli.Context) error {

						fmt.Print("Enter the url you want to add: ")
						var url string
						fmt.Scanln(&url)
						url = strings.TrimSpace(url)

						fmt.Print(`Enter website type (magento|wordpress|shopify|none): `)
						var urlType string
						fmt.Scanln(&urlType)
						urlType = strings.TrimSpace(urlType)

						fmt.Print("Enter website Name: ")
						var websiteName string
						fmt.Scanln(&websiteName)
						websiteName = strings.TrimSpace(websiteName)

						wr := &WebsiteRequest{URL: url, UrlType: urlType, WebsiteName: websiteName}
						bdy, er := json.Marshal(wr)
						if er != nil {
							fmt.Println(er)
							return nil
						}

						res, body, err := gorequest.
							New().
							Post(apiEndPoint+"website").
							Set("Authorization", getToken()).
							Send(string(bdy)).
							End()
						if err != nil {
							fmt.Println(err)
							return nil
						}

						response := parseResponse(body, res)
						if response.Data != nil {
							fmt.Println(response.Data["message"])
						} else {
							fmt.Println("Error: ", response.Error["description"])
						}

						return nil
					},
				},
				{
					Name:  "rm",
					Usage: "permanently remove a website",
					Action: func(c *cli.Context) error {

						var id string
						if c.Args().Len() > 0 {
							id = c.Args().First()
						} else {
							fmt.Print("Enter the id of the website which you want to permanently remove: ")
							fmt.Scanln(&id)
							id = strings.TrimSpace(id)
						}

						if IsValidUUID(id) == false {
							fmt.Println("Please enter a valid website ID. It must be a valid UUID")
							return nil
						}

						fmt.Printf("Going to permanently remove %s website. Are you sure? [Y/n]: ", id)

						var confirm string
						fmt.Scanln(&confirm)

						if strings.ToLower(confirm) == "y" {
							res, body, err := gorequest.
								New().
								Delete(apiEndPoint+"website/"+id).
								Set("Authorization", getToken()).
								End()

							if err != nil {
								fmt.Println(err)
								return nil
							}

							response := parseResponse(body, res)

							if response.Data != nil {
								fmt.Println(response.Data["message"])
							} else {
								fmt.Println("Error: ", response.Error["description"])
							}

						} else {
							fmt.Println("Abort mission!")
						}

						return nil
					},
				},
			},
		},
		{
			Name:    "domain",
			Aliases: []string{"d"},
			Usage:   "options for managing your dexecure domains",
			Subcommands: []*cli.Command{
				{
					Name:  "add",
					Usage: "add a new Dexecure domain",
					Action: func(c *cli.Context) error {
						if getToken() == "" {
							fmt.Println("API token not found please run \"dexecure-cli configure\"")
							return nil
						}
						fmt.Print("Enter the domain you want to optimize: ")
						var origin string
						fmt.Scanln(&origin)
						origin = strings.TrimSpace(origin)

						fmt.Print("Enter Website ID (UUID): ")
						var websiteID string
						fmt.Scanln(&websiteID)
						websiteID = strings.TrimSpace(websiteID)

						if IsValidUUID(websiteID) == false {
							fmt.Println("Please enter a valid domain ID. It must be a valid UUID")
							return nil
						}

						thisDomain := DomainRequest{Origin: origin, WebsiteId: websiteID}

						res, body, err := gorequest.
							New().
							Post(apiEndPoint+"distribution").
							Set("Authorization", getToken()).
							Send(thisDomain).
							End()

						if err != nil {
							fmt.Println(err)
							return nil
						}

						response := parseResponse(body, res)

						if response.Data != nil {
							fmt.Println(response.Data["message"])
						} else {
							fmt.Println("Error: ", response.Error["description"])
						}

						return nil
					},
				},
				{
					Name:  "rm",
					Usage: "permanently removes a domain",
					Action: func(c *cli.Context) error {
						if getToken() == "" {
							fmt.Println("API token not found please run \"dexecure-cli configure\"")
							return nil
						}
						var id string

						if c.Args().Len() > 0 {
							id = c.Args().First()
						} else {
							fmt.Print("Enter the id of the domain which you want to permanently remove: ")
							fmt.Scanln(&id)
							id = strings.TrimSpace(id)
						}

						if IsValidUUID(id) == false {
							fmt.Println("Please enter a valid domain ID. It must be a valid UUID")
							return nil
						}

						fmt.Printf("Going to permanently remove %s domain. Are you sure? [Y/n]: ", id)

						var confirm string
						fmt.Scanln(&confirm)

						if strings.ToLower(confirm) == "y" {

							res, body, err := gorequest.
								New().
								Delete(apiEndPoint+"distribution/"+id).
								Set("Authorization", getToken()).
								End()

							if err != nil {
								fmt.Println(err)
								return nil
							}

							response := parseResponse(body, res)

							if response.Data != nil {
								fmt.Println(response.Data["message"])
							} else {
								fmt.Println("Error: ", response.Error["description"])
							}
						} else {
							fmt.Println("Abort mission!")
						}

						return nil
					},
				},
				{
					Name:  "ls",
					Usage: "Get more information about your domain(s)",
					Subcommands: []*cli.Command{
						{
							Name:  "website",
							Usage: "domain for specific website",
							Action: func(c *cli.Context) error {
								if getToken() == "" {
									fmt.Println("API token not found please run \"dexecure-cli configure\"")
									return nil
								}

								id := ""
								if c.Args().Len() > 0 {
									id = c.Args().First()
									id = strings.TrimSpace(id)
								} else {
									fmt.Print("Enter a Website ID: ")
									fmt.Scanln(&id)
								}
								if IsValidUUID(id) == false {
									fmt.Println("Please enter a valid website ID. It must be a valid UUID")
									return nil
								}

								res, _, err := gorequest.
									New().
									Get(fmt.Sprintf("%sdistribution?websiteId=%s", apiEndPoint, id)).
									Set("Authorization", getToken()).
									End()

								if err != nil {
									fmt.Println(err)
									return nil
								}

								// Read response body
								body, er := ioutil.ReadAll(res.Body)
								if er != nil {
									return nil
								}

								var dr DomainsResponse
								er = json.Unmarshal(body, &dr)
								if er != nil {
									// if expected response Unmarshalling failed then
									// try to Unmarshal error
									er = json.Unmarshal(body, &errorResponse)
									if er != nil {
										return nil
									} else {
										fmt.Println("Error: ", errorResponse.Error.Description)
									}
									return nil
								} else {
									fmt.Println("\nDomains in this website:", len(dr.Data.Distributions))
								}

								if dr.Error.Description != "" {
									fmt.Println(dr.Error.Description)
								} else {
									fmt.Println("")
									for _, domain := range dr.Data.Distributions {
										fmt.Println("-----------------------------------------")
										fmt.Println("Id: ", domain.ID)
										fmt.Println("Origin: ", domain.Origin)
										fmt.Println("Name: ", domain.Name)
										fmt.Println("Type: ", domain.Type)
									}
									fmt.Println("-----------------------------------------")
								}

								return nil
							},
						},
						{
							Name:  "all",
							Usage: "information about your domain(s)",
							Action: func(c *cli.Context) error {

								if getToken() == "" {
									fmt.Println("API token not found please run \"dexecure-cli configure\"")
									return nil
								}

								res, _, err := gorequest.
									New().
									Get(fmt.Sprintf("%sdistribution/", apiEndPoint)).
									Set("Authorization", getToken()).
									End()

								if err != nil {
									fmt.Println(err)
									return nil
								}

								// Read response body
								body, er := ioutil.ReadAll(res.Body)
								if er != nil {
									return nil
								}

								var dr DomainsResponse
								er = json.Unmarshal(body, &dr)
								if er != nil || dr.Error.Code != 0 {
									// if expected response Unmarshalling failed then
									// try to Unmarshal error
									er = json.Unmarshal(body, &errorResponse)
									if er != nil {
										return nil
									} else {
										fmt.Println("Error: ", errorResponse.Error.Description)
									}
									return nil
								} else {
									fmt.Println("\nTotal number of domains:", len(dr.Data.Distributions))
								}

								if dr.Error.Description != "" {
									fmt.Println(dr.Error.Description)
								} else {
									fmt.Println("")
									for _, domain := range dr.Data.Distributions {
										fmt.Println("-----------------------------------------")
										fmt.Println("Id: ", domain.ID)
										fmt.Println("Origin: ", domain.Origin)
										fmt.Println("Website ID: ", domain.WebsiteID)
										fmt.Println("Name: ", domain.Name)
										fmt.Println("Type: ", domain.Type)
									}
									fmt.Println("-----------------------------------------")
								}

								return nil
							},
						},
						{
							Name:  "id",
							Usage: "information about your domain",
							Action: func(c *cli.Context) error {
								if getToken() == "" {
									fmt.Println("API token not found please run \"dexecure-cli configure\"")
									return nil
								}

								id := ""
								if c.Args().Len() > 0 {
									id = c.Args().First()
									id = strings.TrimSpace(id)
								} else {
									fmt.Print("Domain ID: ")
									fmt.Scanln(&id)
								}
								if IsValidUUID(id) == false {
									fmt.Println("Please enter a valid website ID. It must be a valid UUID")
									return nil
								}
								res, _, err := gorequest.
									New().
									Get(fmt.Sprintf("%sdistribution/%s", apiEndPoint, id)).
									Set("Authorization", getToken()).
									End()

								if err != nil {
									fmt.Println(err)
									return nil
								}

								// Read response body
								body, er := ioutil.ReadAll(res.Body)
								if er != nil {
									return nil
								}
								var dr DomainResponse
								er = json.Unmarshal(body, &dr)
								if er != nil || dr.Error.Code != 0 {
									// if expected response Unmarshalling failed then
									// try to Unmarshal error
									er = json.Unmarshal(body, &errorResponse)
									if er != nil {
										return nil
									} else {
										fmt.Println("Error: ", errorResponse.Error.Description)
									}
									return nil
								}

								if dr.Error.Description != "" {
									fmt.Println(dr.Error.Description)
								} else {
									fmt.Println("")
									fmt.Println("-----------------------------------------")
									printDomain(dr.Data)
									fmt.Println("-----------------------------------------")
								}

								return nil
							},
						},
					},
				},
				{
					Name:  "clear",
					Usage: "Clears the cache for a particular domain",
					Action: func(c *cli.Context) error {

						if getToken() == "" {
							fmt.Println("API token not found please run \"dexecure-cli configure\"")
							return nil
						}

						var id string

						if c.Args().Len() > 0 {
							id = c.Args().First()
						} else {
							fmt.Print("Enter the id of the domain whose cache you want to clear: ")
							fmt.Scanln(&id)
							id = strings.TrimSpace(id)
						}

						if IsValidUUID(id) == false {
							fmt.Println("Please enter a valid domain ID. It must be a valid UUID")
							return nil
						}
						fmt.Println("Please choose a option :-")
						fmt.Println("\t1.Clear cache for entire domain")
						fmt.Println("\t2.Clear cache by relative urls(*******/asset/script.js)")
						fmt.Print("How do you want to clean (1/2): ")

						var fc int
						fmt.Scanln(&fc)

						if fc == 1 {
							fmt.Printf("Going to purge the cache for %s domain. Are you sure? [Y/n]: ", id)
							var confirm string
							fmt.Scanln(&confirm)

							if strings.ToLower(confirm) == "y" {

								url := fmt.Sprintf("%sdistribution/%s/clear", apiEndPoint, id)
								res, body, err := gorequest.
									New().
									Post(url).
									Set("Authorization", getToken()).
									Send(`{"url": ["/*"]}`).
									End()

								if err != nil {
									fmt.Println(err)
									return nil
								}

								response := parseResponse(body, res)

								if response.Data != nil {
									fmt.Println(response.Data["message"])
								} else {
									fmt.Println("Error: ", response.Error["description"])
								}

							} else {
								fmt.Println("Abort mission!")
							}
						} else if fc == 2 {
							fmt.Print("Input relative urls(separated by ','): ")
							var urls string
							scanner := bufio.NewScanner(os.Stdin)
							if scanner.Scan() {
								urls = scanner.Text()
							}
							fmt.Printf("\nGoing to purge the cache for %s urls from %s domain. Are you sure? [Y/n]: ", urls, id)

							var confirm string
							fmt.Scanln(&confirm)

							if strings.ToLower(confirm) == "y" {

								urlSlice := strings.Split(urls, ",")
								for i := range urlSlice {
									urlSlice[i] = strings.TrimSpace(urlSlice[i])
								}
								urlB, _ := json.Marshal(urlSlice)

								url := fmt.Sprintf("%sdistribution/%s/clear", apiEndPoint, id)
								res, body, err := gorequest.
									New().
									Post(url).
									Set("Authorization", getToken()).
									Send(fmt.Sprintf(`{"url": %s}`, string(urlB))).
									End()

								if err != nil {
									fmt.Println(err)
									return nil
								}

								response := parseResponse(body, res)

								if response.Data != nil {
									fmt.Println(response.Data["message"])
								} else {
									fmt.Println("Error: ", response.Error["description"])
								}
							} else {
								fmt.Println("Abort mission!")
							}
						}

						return nil
					},
				},
			},
		},
	}

	app.Run(os.Args)
}

func printDomain(dt Data) {
	fmt.Println("Id: ", dt.ID)
	fmt.Println("Origin: ", dt.Origin)
	fmt.Println("Name: ", dt.Name)
	fmt.Println("Type: ", dt.Type)
	fmt.Println("Status: ", dt.Status)
	if dt.WebsiteID != "" {
		fmt.Println("WebsiteID: ", dt.WebsiteID)
	}
	fmt.Println("Region: ", dt.Region)
	fmt.Println("RootPath: ", dt.RootPath)
	fmt.Println("CNames: ", strings.Join(dt.CNames, ", "))
	fmt.Println("JsEnabled: ", dt.JsEnabled)
	fmt.Println("CSSEnabled: ", dt.CSSEnabled)
	fmt.Println("ImageEnabled: ", dt.ImageEnabled)
	fmt.Println("SVGEnabled: ", dt.SVGEnabled)
	fmt.Println("FontEnabled: ", dt.FontEnabled)
	fmt.Println("ProxyEnabled: ", dt.ProxyEnabled)
	fmt.Println("CacheControlImmutable: ", dt.CacheControlImmutable)
	fmt.Println("GIFEnabled: ", dt.GIFEnabled)
	fmt.Println("DefaultCacheTime: ", dt.DefaultCacheTime)
	fmt.Println("Rules: ")
	fmt.Println("*******")
	for _, rule := range dt.Rules {
		fmt.Println("Pattern: ", rule.Pattern)
		fmt.Println("Actions: ", strings.Join(rule.Actions, ", "))
		fmt.Println("*******")
	}
	fmt.Println("AutoResize: ", dt.AutoResize)
	fmt.Println("AutoRotate: ", dt.AutoRotate)
	fmt.Println("HeifEnabled: ", dt.HeifEnabled)
	fmt.Println("TextDetection: ", dt.TextDetection)
	fmt.Println("FaceDetection: ", dt.FaceDetection)
	fmt.Println("Zopflipng: ", dt.Zopflipng)
	fmt.Println("ServerError: ")
	fmt.Println("*******")
	for code, se := range dt.ErrorCaching.ServerError {
		fmt.Println(code, ":", se)
	}
	fmt.Println("*******")
	fmt.Println("ClientError: ")
	fmt.Println("*******")
	fmt.Println("Default: ", dt.ErrorCaching.ClientError.Default)
	fmt.Println("*******")
	fmt.Println("LinkCanonical: ", dt.LinkCanonical)
	fmt.Println("S3BucketIsOrigin: ", dt.S3BucketIsOrigin)
	fmt.Println("S3Bucket Name: ", dt.S3Bucket.Name)
	fmt.Println("S3Bucket Region: ", dt.S3Bucket.Region)

}
