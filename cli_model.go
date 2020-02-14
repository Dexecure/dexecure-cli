package main

import (
	"time"
)

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
		Distributions []Data `json:"distributions"`
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
	Data Data `json:"data"`
}

type UsageResponse struct {
	Status int         `json:"status"`
	Error  struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
		Parameter   string `json:"parameter"`
	} `json:"error"`
	Data   struct {
		Bandwidth     int `json:"bandwidth"`
		Requests      int `json:"requests"`
		Distributions int `json:"distributions"`
	} `json:"data"`
}

type UserResponse struct {
	Status int         `json:"status"`
	Error  struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
		Parameter   string `json:"parameter"`
	} `json:"error"`
	Data   struct {
		ID                      string      `json:"id"`
		FirstName               string      `json:"firstName"`
		LastName                string `json:"lastName"`
		Email                   string      `json:"email"`
		Role                    string      `json:"role"`
		IsEnterprise            int         `json:"isEnterprise"`
		FeaturePrivateS3        int         `json:"featurePrivateS3"`
		FeatureTPO              int         `json:"featureTPO"`
		Coupon                  interface{} `json:"Coupon"`
		IsPaymentDetailsEntered bool        `json:"isPaymentDetailsEntered"`
		IsPasswordEntered       bool        `json:"isPasswordEntered"`
		IsVerified              bool        `json:"isVerified"`
		Plan                    struct {
			ID               string    `json:"id"`
			TeamID           string    `json:"teamId"`
			Tier             int       `json:"tier"`
			Name             string    `json:"name"`
			MaxDistributions int       `json:"max_distributions"`
			MaxBandwidth     int       `json:"max_bandwidth"`
			MaxRequests      int       `json:"max_requests"`
			Price            int       `json:"price"`
			CreatedAt        time.Time `json:"createdAt"`
			UpdatedAt        time.Time `json:"updatedAt"`
		} `json:"Plan"`
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
	Status int         `json:"status"`
	Error  interface{} `json:"error"`
	Data   struct {
		WebsiteURL  string `json:"website_url"`
		WebsiteType string `json:"website_type"`
		WebsiteName string `json:"website_name"`
		ID          string `json:"id"`
	} `json:"data"`
}

type WebsitesResponse struct {
	Status int         `json:"status"`
	Error  struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
		Parameter   string `json:"parameter"`
	} `json:"error"`
	Data   []struct {
		WebsiteURL  string `json:"website_url"`
		WebsiteType string `json:"website_type"`
		WebsiteName string `json:"website_name"`
		ID          string `json:"id"`
	} `json:"data"`
}

type WebsiteRequest struct {
	URL         string `json:"url"`
	UrlType     string `json:"urlType"`
	WebsiteName string `json:"website_name"`
}

type Data struct {
	ID                    string   `json:"id"`
	Origin                string   `json:"origin"`
	Name                  string   `json:"name"`
	Type                  string   `json:"type"`
	Status                string   `json:"status"`
	WebsiteID             string   `json:"websiteId"`
	Region                string   `json:"region"`
	RootPath              string   `json:"rootPath"`
	CNames                []string `json:"CNames"`
	JsEnabled             bool     `json:"jsEnabled"`
	CSSEnabled            bool     `json:"cssEnabled"`
	ImageEnabled          bool     `json:"imageEnabled"`
	SVGEnabled            bool     `json:"SVGEnabled"`
	FontEnabled           bool     `json:"fontEnabled"`
	ProxyEnabled          bool     `json:"proxyEnabled"`
	CacheControlImmutable bool     `json:"cacheControlImmutable"`
	GIFEnabled            bool     `json:"GIFEnabled"`
	DefaultCacheTime      int      `json:"defaultCacheTime"`
	Rules                 []struct {
		Pattern string   `json:"pattern"`
		Actions []string `json:"actions"`
	} `json:"rules"`
	AutoResize    bool `json:"autoResize"`
	AutoRotate    bool `json:"autoRotate"`
	HeifEnabled   bool `json:"heifEnabled"`
	TextDetection bool `json:"textDetection"`
	FaceDetection bool `json:"faceDetection"`
	Zopflipng     bool `json:"zopflipng"`
	ErrorCaching  struct {
		ServerError map[string]interface{} `json:"serverError"`
		ClientError struct {
			Default int `json:"default"`
		} `json:"clientError"`
	} `json:"errorCaching"`
	LinkCanonical    bool `json:"linkCanonical"`
	S3BucketIsOrigin bool `json:"s3BucketIsOrigin"`
	S3Bucket         struct {
		Name   string `json:"name"`
		Region string `json:"region"`
	} `json:"s3Bucket"`
}
