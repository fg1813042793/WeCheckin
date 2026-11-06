package application

type Kind string

const (
	Survey      Kind = "survey"
	Exam        Kind = "exam"
	MaxItems         = 200
	MaxPageSize      = 100
)

func (k Kind) Valid() bool { return k == Survey || k == Exam }

type Settings struct {
	Enabled     bool   `json:"enabled"`
	Title       string `json:"title"`
	Description string `json:"description"`
	LogoURL     string `json:"logoUrl"`
	// Deprecated: accepted and ignored for older configuration clients.
	PublicBaseURL string `json:"publicBaseUrl,omitempty" swaggerignore:"true"`
	ShowEnded     bool   `json:"showEnded"`
	Version       uint64 `json:"version"`
}

type ItemInput struct {
	Kind        Kind   `json:"kind"`
	SourceID    uint   `json:"sourceId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Sort        int    `json:"sort"`
}

type ItemUpdate struct {
	ItemInput
	Version uint64 `json:"version"`
}

type AdminItem struct {
	ItemInput
	ID                uint64 `json:"id"`
	Version           uint64 `json:"version"`
	ManagementTitle   string `json:"managementTitle"`
	Category          string `json:"category"`
	DisplayTitle      string `json:"displayTitle"`
	Available         bool   `json:"available"`
	UnavailableReason string `json:"unavailableReason"`
	CanManage         bool   `json:"canManage"`
}

type AdminResponse struct {
	Settings Settings    `json:"settings"`
	Items    []AdminItem `json:"items"`
}

// PublicResponse deliberately contains no embedded persistence/source models.
type PublicResponse struct {
	Enabled     bool         `json:"enabled"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	LogoURL     string       `json:"logoUrl"`
	Items       []PublicItem `json:"items"`
}

type PublicItem struct {
	ID               uint64 `json:"id"`
	Kind             Kind   `json:"kind"`
	SourceID         uint   `json:"sourceId"`
	Title            string `json:"title"`
	ManagementTitle  string `json:"managementTitle"`
	Category         string `json:"category"`
	Description      string `json:"description"`
	StartTime        int64  `json:"startTime"`
	EndTime          int64  `json:"endTime"`
	State            string `json:"state"`
	LoginRequired    bool   `json:"loginRequired"`
	PasswordRequired bool   `json:"passwordRequired"`
	// URL is a root-relative filling route; the browser resolves it on its current site.
	URL string `json:"url"`
}

type OptionsQuery struct {
	Kind     Kind
	Keyword  string
	Page     int
	PageSize int
}

type SourceOption struct {
	Kind             Kind   `json:"kind"`
	SourceID         uint   `json:"sourceId"`
	ManagementTitle  string `json:"managementTitle"`
	Category         string `json:"category"`
	DisplayTitle     string `json:"displayTitle"`
	StartTime        int64  `json:"startTime"`
	EndTime          int64  `json:"endTime"`
	LoginRequired    bool   `json:"loginRequired"`
	PasswordRequired bool   `json:"passwordRequired"`
}

type OptionsResponse struct {
	List     []SourceOption `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

type DeleteResponse struct {
	ID uint64 `json:"id"`
}
