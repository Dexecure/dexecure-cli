package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/kyokomi/emoji"
	"github.com/mitchellh/mapstructure"
	"github.com/parnurzeal/gorequest"
	"github.com/tucnak/store"
	"gopkg.in/urfave/cli.v1"
)

var apiEndPoint = "http://localhost:8080/api/v1/"

type User struct {
	Id       string `json:"id"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Distribution struct {
	Id     string  `json:"id"`
	Origin string  `json:"origin"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Usage  float64 `json:"usage"`
	Status string  `json:"status"`
}

type DistributionList []Distribution

type TokenSettings struct {
	Token string
}

type Response struct {
	Data  map[string]interface{}
	Error map[string]interface{}
}

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

func main() {
	// var token string = ""
	app := cli.NewApp()

	app.Name = "Dexecure CLI"
	app.Usage = "Interact with your Dexecure account"
	app.Version = "0.0.1"
	app.Copyright = "Dexecure PTE LTD."
	app.EnableBashCompletion = true

	// config management
	store.Init("dexecure")

	app.Commands = []cli.Command{
		{
			Name:    "login",
			Aliases: []string{"a"},
			Usage:   "Login using your Dexecure credentials",
			Action: func(c *cli.Context) error {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("enter your email: ")
				email, _ := reader.ReadString('\n')
				email = strings.TrimRight(email, "\n")

				fmt.Print("enter your password:")
				emoji.Print(":key: ")
				passwordBytes, _ := gopass.GetPasswd()
				password := string(passwordBytes[:])

				thisUser := User{Email: email, Password: password}

				res, body, err := gorequest.
					New().
					Post(apiEndPoint + "user/login").
					Send(thisUser).
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
			Usage:   "Total bandwidth served via Dexecure across all your distributions",
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
			Name:    "distribution",
			Aliases: []string{"d"},
			Usage:   "options for managing your dexecure distributions",
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "add a new Dexecure distribution",
					Action: func(c *cli.Context) error {

						fmt.Print("Enter the domain you want to optimize - ")
						reader := bufio.NewReader(os.Stdin)
						origin, _ := reader.ReadString('\n')
						origin = strings.TrimRight(origin, "\n")

						thisDistribution := Distribution{Origin: origin}

						res, body, err := gorequest.
							New().
							Post(apiEndPoint+"distribution").
							Set("Authorization", getToken()).
							Send(thisDistribution).
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
					Usage: "permanently removes a distribution",
					Action: func(c *cli.Context) error {

						var id string
						reader := bufio.NewReader(os.Stdin)

						if len(c.Args()) > 0 {
							id = c.Args().First()
						} else {
							fmt.Print("Enter the id of the distribution which you want to permanently remove - ")
							id, _ = reader.ReadString('\n')
							id = strings.TrimRight(id, "\n")
						}

						if IsValidUUID(id) == false {
							fmt.Println("Please enter a valid distribution ID. It must be a valid UUID")
							return nil
						}

						fmt.Printf("Going to permanently remove %s distribution. Are you sure? [Y/n] ", id)
						confirm, _ := reader.ReadByte()

						if confirm == 'Y' {

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
					Usage: "Get more information about your distribution(s)",
					Action: func(c *cli.Context) error {

						if len(c.Args()) > 0 {
							// printing information about one particular distribution
							id := c.Args().First()

							url := fmt.Sprintf("%sdistribution/%s", apiEndPoint, id)

							if IsValidUUID(id) == false {
								fmt.Println("Please enter a valid distribution ID. It must be a valid UUID")
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
								distributionMap := response.Data["distributions"]
								var distribution Distribution
								mapstructure.Decode(distributionMap, &distribution)

								fmt.Println("Id: ", distribution.Id)
								fmt.Println("Origin: ", distribution.Origin)
								fmt.Println("Name: ", distribution.Name)
								fmt.Println("Type: ", distribution.Type)
								fmt.Println("Status: ", distribution.Status)
								fmt.Printf("Usage: %.2f MB \n", distribution.Usage/(1024*1024))
							} else {
								fmt.Println("Error: ", response.Error["description"])
							}

						} else {
							// printing list of all your distributions
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
								var distributions DistributionList
								mapstructure.Decode(response.Data["distributions"], &distributions)

								fmt.Println("Total number of distributions: ", len(distributions), "\n")
								for _, distribution := range distributions {
									fmt.Println("Id: ", distribution.Id)
									fmt.Println("Origin: ", distribution.Origin)
									fmt.Println("Name: ", distribution.Name)
									fmt.Println("Type: ", distribution.Type)
									fmt.Println("Status: ", distribution.Status)
									fmt.Printf("Usage: %.2f MB \n", distribution.Usage/(1024*1024))
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
					Usage: "Clears the cahce for a particular distribution",
					Action: func(c *cli.Context) error {

						var id string
						reader := bufio.NewReader(os.Stdin)

						if len(c.Args()) > 0 {
							id = c.Args().First()
						} else {
							fmt.Print("Enter the id of the distribution whose cache you want to clear - ")
							id, _ = reader.ReadString('\n')
							id = strings.TrimRight(id, "\n")
						}

						if IsValidUUID(id) == false {
							fmt.Println("Please enter a valid distribution ID. It must be a valid UUID")
							return nil
						}

						fmt.Printf("Going to delete the cache for %s distribution. Are you sure? [Y/n] ", id)
						confirm, _ := reader.ReadByte()

						if confirm == 'Y' {

							url := fmt.Sprintf("%sdistribution/%s/clear", apiEndPoint, id)
							res, body, err := gorequest.
								New().
								Post(url).
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
					Name:  "enable",
					Usage: "Enables a previously disabled Dexecure distribution",
					Action: func(c *cli.Context) error {

						var id string
						reader := bufio.NewReader(os.Stdin)

						if len(c.Args()) > 0 {
							id = c.Args().First()
						} else {
							fmt.Print("Enter the id of the distribution which you want to enable - ")
							id, _ = reader.ReadString('\n')
							id = strings.TrimRight(id, "\n")
						}

						if IsValidUUID(id) == false {
							fmt.Println("Please enter a valid distribution ID. It must be a valid UUID")
							return nil
						}

						fmt.Printf("Going to enable %s distribution. Are you sure? [Y/n] ", id)
						confirm, _ := reader.ReadByte()

						if confirm == 'Y' {

							url := fmt.Sprintf("%sdistribution/%s/enable", apiEndPoint, id)
							res, body, err := gorequest.
								New().
								Post(url).
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
					Name:  "disable",
					Usage: "Disables a Dexecure distribution",
					Action: func(c *cli.Context) error {

						var id string
						reader := bufio.NewReader(os.Stdin)

						if len(c.Args()) > 0 {
							id = c.Args().First()
						} else {
							fmt.Print("Enter the id of the distribution which you want to disable - ")
							id, _ = reader.ReadString('\n')
							id = strings.TrimRight(id, "\n")
						}

						if IsValidUUID(id) == false {
							fmt.Println("Please enter a valid distribution ID. It must be a valid UUID")
							return nil
						}

						fmt.Printf("Going to disable %s distribution. Are you sure? [Y/n] ", id)
						confirm, _ := reader.ReadByte()

						if confirm == 'Y' {

							url := fmt.Sprintf("%sdistribution/%s/disable", apiEndPoint, id)
							res, body, err := gorequest.
								New().
								Post(url).
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
	}

	app.Run(os.Args)
}
