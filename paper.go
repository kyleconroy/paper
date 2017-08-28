package blog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func NewClient(token string) *APIClient {
	return &APIClient{
		Token: token,
		HTTP:  http.Client{},
	}
}

type Client interface {
	ListDocs(context.Context, *ListPaperDocsArgs) (*ListPaperDocsResponse, error)
	DownloadDoc(context.Context, *PaperDocExport) (*PaperDocExportResult, []byte, error)
	GetDocFolderInfo(context.Context, *RefPaperDoc) (*FoldersContainingPaperDoc, error)
}

type APIClient struct {
	Token string
	HTTP  http.Client
}

type APIError struct {
	Summary  string            `json:"error_summary"`
	Metadata map[string]string `json:"error"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s: %q", e.Summary, e.Metadata)
}

func (c *APIClient) rpc(ctx context.Context, url string, in interface{}, out interface{}) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var apierr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apierr); err != nil {
			return err
		}
		return apierr
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *APIClient) content(ctx context.Context, url string, in interface{}, out interface{}) ([]byte, error) {
	var contents []byte
	body, err := json.Marshal(in)
	if err != nil {
		return contents, err
	}
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Dropbox-API-Arg", string(body))
	resp, err := c.HTTP.Do(req.WithContext(ctx))
	if err != nil {
		return contents, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apierr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apierr); err != nil {
			return contents, err
		}
		return contents, apierr
	}

	if result := resp.Header.Get("Dropbox-API-Result"); result != "" {
		if err := json.Unmarshal([]byte(result), out); err != nil {
			return contents, err
		}
	}

	return ioutil.ReadAll(resp.Body)
}

type ListPaperDocsFilterBy string

const (
	ListPaperDocsFilterByAccessed ListPaperDocsFilterBy = "accessed"
	ListPaperDocsFilterByModified                       = "modified"
	ListPaperDocsFilterByCreated                        = "created"
)

type ListPaperDocsSortBy string

const (
	ListPaperDocsSortByAccessed ListPaperDocsSortBy = "accessed"
	ListPaperDocsSortByModified                     = "modified"
	ListPaperDocsSortByCreated                      = "created"
)

type ListPaperDocsSortOrder string

const (
	ListPaperDocsSortOrderAsc  ListPaperDocsSortOrder = "ascending"
	ListPaperDocsSortOrderDesc                        = "descending"
)

type ListPaperDocsArgs struct {
	FilterBy  ListPaperDocsFilterBy  `json:"filter_by,omitempty"`
	SortBy    ListPaperDocsSortBy    `json:"sort_by,omitempty"`
	SortOrder ListPaperDocsSortOrder `json:"sort_order,omitempty"`
	Limit     int32                  `json:"limit,omitempty"`
}

type Cursor struct {
	Value      string `json:"value"`
	Expiration string `json:"expiration"` // TODO: Make a time.Time
}

type ListPaperDocsResponse struct {
	DocIDs  []string `json:"doc_ids"`
	Cursor  Cursor   `json:"cursor"`
	HasMore bool     `json"has_more"`
}

func (c *APIClient) ListDocs(ctx context.Context, in *ListPaperDocsArgs) (*ListPaperDocsResponse, error) {
	var out ListPaperDocsResponse
	return &out, c.rpc(ctx, "https://api.dropboxapi.com/2/paper/docs/list", in, &out)
}

type ExportFormat string

const (
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatHTML     ExportFormat = "html"
)

type PaperDocExport struct {
	DocID  string       `json:"doc_id,omitempty"`
	Format ExportFormat `json:"export_format,omitempty"`
}

type PaperDocExportResult struct {
	Owner    string `json:"owner"`
	Title    string `json:"title"`
	Revision int64  `json:"revision"`
	MIME     string `json:"mime_type"`
}

func (c *APIClient) DownloadDoc(ctx context.Context, in *PaperDocExport) (*PaperDocExportResult, []byte, error) {
	var out PaperDocExportResult
	blob, err := c.content(ctx, "https://api.dropboxapi.com/2/paper/docs/download", in, &out)
	return &out, blob, err
}

type RefPaperDoc struct {
	DocID string `json:"doc_id"`
}

type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FolderSharingPolicyType string

const (
	FolderSharingPolicyTeam       FolderSharingPolicyType = "team"
	FolderSharingPolicyInviteOnly                         = "invite_only"
)

type FoldersContainingPaperDoc struct {
	FolderSharingPolicyType FolderSharingPolicyType
	Folders                 []Folder
}

func (c *APIClient) GetDocFolderInfo(ctx context.Context, in *RefPaperDoc) (*FoldersContainingPaperDoc, error) {
	var out FoldersContainingPaperDoc
	return &out, c.rpc(ctx, "https://api.dropboxapi.com/2/paper/docs/get_folder_info", in, &out)
}

var _ Client = &APIClient{}
