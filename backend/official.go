package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

const officialAPIURL = "https://api22-normal-c-alisg.tiktokv.com/aweme/v1/feed/"

const mobileUA = "com.zhiliaoapp.musically/2023501030 (Linux; U; Android 13; en; " +
	"Pixel 7; Build/TQ2A.230505.002; Cronet/58.0.2991.0)"

// MediaInfo is the resolved media for a single post (official API).
type MediaInfo struct {
	MediaID       string
	Username      string
	VideoURL      string
	ImageURLs     []string
	AvatarURLs    []string
	GifAvatarURLs []string
}

func (m MediaInfo) IsImagePost() bool {
	return m.VideoURL == "" && len(m.ImageURLs) > 0
}

// apiResponse mirrors the JSON shape of the feed endpoint we read.
type apiResponse struct {
	AwemeList []struct {
		AwemeID string `json:"aweme_id"`
		Video   struct {
			PlayAddr     urlList `json:"play_addr"`
			DownloadAddr urlList `json:"download_addr"`
		} `json:"video"`
		ImagePostInfo struct {
			Images []struct {
				DisplayImage urlList `json:"display_image"`
			} `json:"images"`
		} `json:"image_post_info"`
		Author struct {
			UniqueID     string  `json:"unique_id"`
			AvatarMedium urlList `json:"avatar_medium"`
			VideoIcon    urlList `json:"video_icon"`
		} `json:"author"`
	} `json:"aweme_list"`
}

type urlList struct {
	URLList []string `json:"url_list"`
}

func (u urlList) first() string {
	if len(u.URLList) > 0 {
		return u.URLList[0]
	}
	return ""
}

func officialParams(mediaID string) url.Values {
	v := url.Values{}
	v.Set("aweme_id", mediaID)
	v.Set("iid", "7238789370386695942")
	v.Set("device_id", "7238787983025079814")
	v.Set("resolution", "1080*2400")
	v.Set("channel", "googleplay")
	v.Set("app_name", "musical_ly")
	v.Set("version_code", "350103")
	v.Set("device_platform", "android")
	v.Set("device_type", "Pixel 7")
	v.Set("os_version", "13")
	return v
}

// GetMedia fetches media metadata for a single post from the official feed API.
// Returns nil if the post can't be resolved. Retries on HTTP 429 with a 5s delay.
func GetMedia(client *http.Client, mediaID string, watermark bool) (*MediaInfo, error) {
	const max429 = 3
	endpoint := officialAPIURL + "?" + officialParams(mediaID).Encode()

	for attempt := 0; attempt <= max429; attempt++ {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", mobileUA)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			if attempt < max429 {
				time.Sleep(5 * time.Second)
				continue
			}
			return nil, nil
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if len(body) == 0 {
			return nil, nil
		}

		var data apiResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, err
		}
		if len(data.AwemeList) == 0 {
			return nil, nil
		}

		aweme := data.AwemeList[0]
		if aweme.AwemeID != mediaID {
			return nil, nil
		}

		info := &MediaInfo{
			MediaID:       mediaID,
			Username:      aweme.Author.UniqueID,
			AvatarURLs:    aweme.Author.AvatarMedium.URLList,
			GifAvatarURLs: aweme.Author.VideoIcon.URLList,
		}
		if watermark {
			info.VideoURL = aweme.Video.DownloadAddr.first()
		} else {
			info.VideoURL = aweme.Video.PlayAddr.first()
		}
		for _, img := range aweme.ImagePostInfo.Images {
			if u := img.DisplayImage.first(); u != "" {
				info.ImageURLs = append(info.ImageURLs, u)
			}
		}
		return info, nil
	}
	return nil, nil
}
