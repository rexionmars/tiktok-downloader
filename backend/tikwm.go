package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	hdImageEndpoint = "https://www.tikwm.com/api/"
	hdVideoSubmit   = "https://www.tikwm.com/api/video/task/submit"
	hdVideoResult   = "https://www.tikwm.com/api/video/task/result"
)

// HdResult holds the HD media resolved from tikwm.com.
type HdResult struct {
	MediaID   string
	Username  string
	VideoURL  string
	ImageURLs []string
}

// GetHdImages fetches HD image URLs for a photo post via tikwm.com.
func GetHdImages(client *http.Client, mediaIDOrURL, fallbackUser string) (*HdResult, error) {
	endpoint := hdImageEndpoint + "?url=" + url.QueryEscape(mediaIDOrURL) + "&hd=1"
	req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", desktopUA)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var parsed struct {
		Code int `json:"code"`
		Data struct {
			Play   string   `json:"play"`
			Images []string `json:"images"`
			Author struct {
				UniqueID string `json:"unique_id"`
			} `json:"author"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 0 {
		return nil, fmt.Errorf("tikwm.com error for %s", mediaIDOrURL)
	}

	username := parsed.Data.Author.UniqueID
	if username == "" {
		username = fallbackUser
	}
	return &HdResult{
		MediaID:   mediaIDOrURL,
		Username:  username,
		VideoURL:  parsed.Data.Play,
		ImageURLs: parsed.Data.Images,
	}, nil
}

// GetHdVideo submits and polls the tikwm.com task endpoint for an HD video.
func GetHdVideo(client *http.Client, mediaIDOrURL, fallbackUser string) (*HdResult, error) {
	form := url.Values{}
	form.Set("url", mediaIDOrURL)
	form.Set("web", "1")

	req, _ := http.NewRequest(http.MethodPost, hdVideoSubmit, strings.NewReader(form.Encode()))
	req.Header.Set("User-Agent", desktopUA)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var submit struct {
		Code int `json:"code"`
		Data struct {
			TaskID string `json:"task_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &submit); err != nil {
		return nil, err
	}
	if submit.Code != 0 || submit.Data.TaskID == "" {
		return nil, fmt.Errorf("failed to submit HD video task for %s", mediaIDOrURL)
	}

	const maxRetries = 15
	resultURL := hdVideoResult + "?task_id=" + url.QueryEscape(submit.Data.TaskID)
	for i := 0; i < maxRetries; i++ {
		rreq, _ := http.NewRequest(http.MethodGet, resultURL, nil)
		rreq.Header.Set("User-Agent", desktopUA)
		rresp, err := client.Do(rreq)
		if err != nil {
			return nil, err
		}
		rbody, _ := io.ReadAll(rresp.Body)
		rresp.Body.Close()

		var result struct {
			Code int `json:"code"`
			Data struct {
				Status int `json:"status"`
				Detail struct {
					PlayURL string      `json:"play_url"`
					Size    json.Number `json:"size"`
					Author  struct {
						UniqueID string `json:"unique_id"`
					} `json:"author"`
				} `json:"detail"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rbody, &result); err == nil && result.Code == 0 {
			size, _ := strconv.ParseInt(result.Data.Detail.Size.String(), 10, 64)
			if result.Data.Status == 2 && size > 0 {
				username := result.Data.Detail.Author.UniqueID
				if username == "" {
					username = fallbackUser
				}
				return &HdResult{
					MediaID:  mediaIDOrURL,
					Username: username,
					VideoURL: result.Data.Detail.PlayURL,
				}, nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, fmt.Errorf("HD video task for %s not ready after %d attempts", mediaIDOrURL, maxRetries)
}
