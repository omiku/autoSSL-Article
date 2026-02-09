package onepanelsite

import "time"

type apiResponse interface {
	GetCode() int32
	GetMessage() string
}

type apiResponseBase struct {
	Code    *int32  `json:"code,omitempty"`
	Message *string `json:"message,omitempty"`
}

func (r *apiResponseBase) GetCode() int32 {
	if r.Code == nil {
		return 0
	}
	return *r.Code
}

func (r *apiResponseBase) GetMessage() string {
	if r.Message == nil {
		return ""
	}
	return *r.Message
}

type GetHttpsConfResponse struct {
	apiResponseBase

	Data *struct {
		Enable      bool     `json:"enable"`
		HttpConfig  string   `json:"httpConfig"`
		SSLProtocol []string `json:"SSLProtocol"`
		Algorithm   string   `json:"algorithm"`
		Hsts        bool     `json:"hsts"`
	} `json:"data,omitempty"`
}

type UploadWebsiteSSLRequest struct {
	SSLID           int64  `json:"sslID"`
	Type            string `json:"type"`
	Certificate     string `json:"certificate"`
	CertificatePath string `json:"certificatePath"`
	PrivateKey      string `json:"privateKey"`
	PrivateKeyPath  string `json:"privateKeyPath"`
	Description     string `json:"description"`
}

type UploadWebsiteSSLResponse struct {
	apiResponseBase
}

type SearchWebsiteSSLRequest struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"pageSize"`
}

type SearchWebsiteRequest struct {
	Name           string `json:"name"`
	Page           int32  `json:"page"`
	PageSize       int32  `json:"pageSize"`
	OrderBy        string `json:"orderBy"`
	Order          string `json:"order"`
	WebsiteGroupID int32  `json:"websiteGroupId"`
	Type           string `json:"type"`
}

// SearchWebsiteResponse 搜索网站响应
type SearchWebsiteResponse struct {
	apiResponseBase
	Data *struct {
		Total int32 `json:"total"`
		Items []*struct {
			ID            int64       `json:"id"`
			CreatedAt     time.Time   `json:"createdAt"`
			Protocol      string      `json:"protocol"`
			PrimaryDomain string      `json:"primaryDomain"`
			Type          string      `json:"type"`
			Alias         string      `json:"alias"`
			Remark        string      `json:"remark"`
			Status        string      `json:"status"`
			ExpireDate    time.Time   `json:"expireDate"`
			SitePath      string      `json:"sitePath"`
			AppName       string      `json:"appName"`
			RuntimeName   string      `json:"runtimeName"`
			SslExpireDate time.Time   `json:"sslExpireDate"`
			SslStatus     string      `json:"sslStatus"`
			AppInstallID  int         `json:"appInstallId"`
			ChildSites    interface{} `json:"childSites"`
			RuntimeType   string      `json:"runtimeType"`
			Favorite      bool        `json:"favorite"`
			IPV6          bool        `json:"IPV6"`
		} `json:"items"`
	} `json:"data,omitempty"`
}

type SearchWebsiteSSLResponse struct {
	apiResponseBase

	Data *struct {
		Items []*struct {
			ID          int64  `json:"id"`
			PEM         string `json:"pem"`
			PrivateKey  string `json:"privateKey"`
			Domains     string `json:"domains"`
			Description string `json:"description"`
			Status      string `json:"status"`
			UpdatedAt   string `json:"updatedAt"`
			CreatedAt   string `json:"createdAt"`
		} `json:"items"`
		Total int32 `json:"total"`
	} `json:"data,omitempty"`
}

type UpdateCoreSettingsSSLRequest struct {
	Cert        string `json:"cert"`
	Key         string `json:"key"`
	SSLType     string `json:"sslType"`
	SSL         string `json:"ssl"`
	SSLID       int64  `json:"sslID"`
	AutoRestart string `json:"autoRestart"`
}

type UpdateCoreSettingsSSLResponse struct {
	apiResponseBase
}

type UpdateHttpsConfRequest struct {
	WebsiteID       int64    `json:"websiteId"`
	Enable          bool     `json:"enable"`
	Type            string   `json:"type"`
	WebsiteSSLID    int64    `json:"websiteSSLId"`
	PrivateKey      string   `json:"privateKey"`
	Certificate     string   `json:"certificate"`
	PrivateKeyPath  string   `json:"privateKeyPath"`
	CertificatePath string   `json:"certificatePath"`
	ImportType      string   `json:"importType"`
	HttpConfig      string   `json:"httpConfig"`
	SSLProtocol     []string `json:"SSLProtocol"`
	Algorithm       string   `json:"algorithm"`
	Hsts            bool     `json:"hsts"`
}

type UpdateHttpsConfResponse struct {
	apiResponseBase
}

type GetWebsiteSSLResponse struct {
	apiResponseBase

	Data *struct {
		ID            int64  `json:"id"`
		Provider      string `json:"provider"`
		Description   string `json:"description"`
		PrimaryDomain string `json:"primaryDomain"`
		Domains       string `json:"domains"`
		Type          string `json:"type"`
		Organization  string `json:"organization"`
		Status        string `json:"status"`
		StartDate     string `json:"startDate"`
		ExpireDate    string `json:"expireDate"`
		CreatedAt     string `json:"createdAt"`
		UpdatedAt     string `json:"updatedAt"`
	} `json:"data,omitempty"`
}
