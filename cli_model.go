package main

type User struct {
	Id       string `json:"id"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

/*type Domain struct {
	Id        string  `json:"id"`
	WebsiteId string  `json:"websiteId"`
	Origin    string  `json:"origin"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Usage     float64 `json:"usage"`
}*/

type DomainsResponse struct {
	Status int `json:"status"`
	Error  struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
		Parameter   string `json:"parameter"`
	} `json:"error"`
	Data struct {
		Distributions []struct {
			ID        string  `json:"id"`
			Origin    string  `json:"origin"`
			Name      string  `json:"name"`
			Type      string  `json:"type"`
			Usage     float64 `json:"usage"`
			WebsiteID string  `json:"websiteId"`
		} `json:"distributions"`
	} `json:"data"`
}

type DomainRequest struct {
	WebsiteId string `json:"websiteId"`
	Origin    string `json:"origin"`
}

type DomainResponse struct {
	Status int `json:"status"`
	Error  struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
		Parameter   string `json:"parameter"`
	} `json:"error"`
	Data struct {
		ID     string  `json:"id"`
		Origin string  `json:"origin"`
		Name   string  `json:"name"`
		Type   string  `json:"type"`
		Usage  float64 `json:"usage"`
	} `json:"data"`
}

type TokenSettings struct {
	Token string
}

type Response struct {
	Data  map[string]interface{}
	Error map[string]interface{}
}

type ErrorResponse struct {
	Status int `json:"status"`
	Error  struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
		Parameter   string `json:"parameter"`
	} `json:"error"`
	Data struct {
	} `json:"data"`
}

type WebsiteResponse struct {
	Status int `json:"status"`
	Error  struct {
	} `json:"error"`
	Data []struct {
		WebsiteURL  string      `json:"website_url"`
		WebsiteType string      `json:"website_type"`
		WebsiteName interface{} `json:"website_name"`
		ID          string      `json:"id"`
	} `json:"data"`
}

type WebsiteRequest struct {
	URL         string `json:"url"`
	UrlType     string `json:"urlType"`
	WebsiteName string `json:"website_name"`
}
