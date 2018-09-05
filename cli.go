package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/kyokomi/emoji"
	"github.com/mitchellh/mapstructure"
	"github.com/parnurzeal/gorequest"
	"github.com/tucnak/store"
	"gopkg.in/urfave/cli.v2"
)

var apiEndPoint = "https://dao-dev.dexecure.com/api/v1/"
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
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
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
		if responseJSON["error"].(map[string]interface{})["error"] != nil {
			response.Error = responseJSON["error"].(map[string]interface{})["error"].(map[string]interface{})
			return response
		}

		responseStatus := responseJSON["status"].(float64)

		if responseStatus == 200 {
			response.Data = responseJSON["data"].(map[string]interface{})
		} else {
			response.Error = responseJSON["error"].(map[string]interface{})
		}
	} else {
		fmt.Println("Request to the API failed")
	}

	return response
}

func credentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Email: ")
	email, _ := reader.ReadString('\n')

	fmt.Print("Enter Password:")
	emoji.Print(":key: ")
	password, _ := gopass.GetPasswdMasked()

	return strings.TrimSpace(email), string(password[:])
}

func main() {
	// var token string = ""
	app := &cli.App{}

	app.Name = "Dexecure CLI"
	app.Usage = "Interact with your Dexecure account"
	app.Version = "0.0.1"
	app.Copyright = "Dexecure PTE LTD."
	app.EnableShellCompletion = true

	// config management
	store.Init("dexecure")

	app.Commands = []*cli.Command{
		{
			Name:    "login",
			Aliases: []string{"a"},
			Usage:   "Login using your Dexecure credentials",
			Action: func(c *cli.Context) error {
				email, password := credentials()
				thisUser := User{Email: email, Password: password}

				res, body, err := gorequest.
					New().
					Post(apiEndPoint + "user/login").
					Send(fmt.Sprint(`{"email":"`, thisUser.Email, `", "password":"`, thisUser.Password, `"}`)).
					End()

				if err != nil {
					fmt.Println(err)
					return nil
				}

				response := parseResponse(body, res)

				if response.Data != nil {
					token := response.Data["token"].(string)
					saveToken(token)
					fmt.Println("you have been logged in successfully,", email)
				} else {
					fmt.Println("Error: ", response.Error["description"])
				}

				return nil
			},
		},
		{
			Name:    "logout",
			Aliases: []string{"l"},
			Usage:   "Logout of your current session",
			Action: func(c *cli.Context) error {
				saveToken("")
				fmt.Println("You have been logged out.")
				return nil
			},
		},
		{
			Name:    "usage",
			Aliases: []string{"l"},
			Usage:   "Total bandwidth served via Dexecure across all your domains",
			Action: func(c *cli.Context) error {
				res, body, err := gorequest.
					New().
					Get(apiEndPoint+"user/usage").
					Set("Authorization", getToken()).
					End()

				if err != nil {
					fmt.Println(err)
					return nil
				}

				response := parseResponse(body, res)

				if response.Data != nil {
					fmt.Printf("You have used %.2f MB this month \n", response.Data["usage"].(float64)/(1024*1024))
				} else {
					fmt.Println("Error: ", response.Error["description"])
				}

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
					Action: func(c *cli.Context) error {

						id := ""
						if c.Args().Len() > 0 {
							id = c.Args().First()
							id = strings.TrimSpace(id)
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

						fmt.Println("\nTotal number of websites :", len(wr.Data))
						for _, website := range wr.Data {
							fmt.Println("-----------------------------------------")
							fmt.Println("ID: ", website.ID)
							fmt.Println("Website URL: ", website.WebsiteURL)
							fmt.Println("Website Type: ", website.WebsiteType)
							fmt.Println("Website Name: ", website.WebsiteName)
						}

						return nil
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
							Post(apiEndPoint+"website/").
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

						thisDomain := Domain{Origin: origin, WebsiteId: websiteID}

						res, body, err := gorequest.
							New().
							Post(apiEndPoint+"distribution/").
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
					Action: func(c *cli.Context) error {

						if c.Args().Len() > 0 {
							// printing information about one particular domain
							id := c.Args().First()

							url := fmt.Sprintf("%sdomain/%s", apiEndPoint, id)

							if IsValidUUID(id) == false {
								fmt.Println("Please enter a valid domain ID. It must be a valid UUID")
								return nil
							}

							res, body, err := gorequest.
								New().
								Get(url).
								Set("Authorization", getToken()).
								End()

							if err != nil {
								fmt.Println(err)
								return nil
							}

							response := parseResponse(body, res)

							if response.Data != nil {
								domainMap := response.Data["distributions"]
								var domain Domain
								mapstructure.Decode(domainMap, &domain)

								fmt.Println("Id: ", domain.Id)
								fmt.Println("Origin: ", domain.Origin)
								fmt.Println("Name: ", domain.Name)
								fmt.Println("Type: ", domain.Type)
								fmt.Println("Status: ", domain.Status)
								fmt.Printf("Usage: %.2f MB \n", domain.Usage/(1024*1024))
							} else {
								fmt.Println("Error: ", response.Error["description"])
							}

						} else {
							// printing list of all your domains
							res, body, err := gorequest.
								New().
								Get(apiEndPoint+"distribution").
								Set("Authorization", getToken()).
								End()

							if err != nil {
								fmt.Println(err)
								return nil
							}

							response := parseResponse(body, res)

							if response.Data != nil {
								var domains DomainList
								mapstructure.Decode(response.Data["distributions"], &domains)

								fmt.Println("Total number of domains: ", len(domains), "\n")
								for _, domain := range domains {
									fmt.Println("Id: ", domain.Id)
									fmt.Println("Origin: ", domain.Origin)
									fmt.Println("Name: ", domain.Name)
									fmt.Println("Type: ", domain.Type)
									fmt.Println("Status: ", domain.Status)
									fmt.Printf("Usage: %.2f MB \n", domain.Usage/(1024*1024))
									fmt.Println("")
								}
							} else {
								fmt.Println("Error: ", response.Error["description"])
							}

						}

						return nil
					},
				},
				{
					Name:  "clear",
					Usage: "Clears the cache for a particular domain",
					Action: func(c *cli.Context) error {

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

						fmt.Printf("Going to purge the cache for %s domain. Are you sure? [Y/n]: ", id)
						var confirm string
						fmt.Scanln(&confirm)

						if strings.ToLower(confirm) == "y" {

							url := fmt.Sprintf("%sdomain/%s/clear", apiEndPoint, id)
							res, body, err := gorequest.
								New().
								Post(url).
								Set("Authorization", getToken()).
								Send(`{"url": "*"}`).
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
	}

	app.Run(os.Args)
}
