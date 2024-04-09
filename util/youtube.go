package util

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Youtube struct {
	VideoID    string
	client     *req.Client
	jsonResult gjson.Result
}

var ErrNoYtCaptions = errors.New("cannot find youtube captions")

func NewYoutube(url string, proxy string) (*Youtube, error) {
	_, client, err := MakeHTTPClient(proxy, 15*time.Second)
	if err != nil {
		return nil, err
	}
	client.SetRedirectPolicy(req.NoRedirectPolicy())
	if strings.HasPrefix(url, "https://youtu.be") {
		resp, err := client.R().Head(url)
		if err != nil {
			return nil, err
		}
		if resp.IsErrorState() {
			return nil, errors.New("cannot head " + url + ", status: " + strconv.Itoa(resp.GetStatusCode()))
		}
		url = resp.GetHeader("location")
	}
	if !strings.HasPrefix(url, "https://www.youtube.com") {
		return nil, errors.New("not a valid youtube link: " + url)
	}
	arr := regexp.MustCompile("v=([^&]+)").FindStringSubmatch(url)
	if len(arr) == 0 {
		return nil, errors.New("invalid youtube video url: " + url)
	}
	resp, err := client.R().Get("https://www.youtube.com/watch?v=" + arr[1])
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, errors.New("cannot fetch youtube url: " + strconv.Itoa(resp.GetStatusCode()))
	}
	arr = regexp.MustCompile("var ytInitialPlayerResponse = (.*?);var meta =").FindStringSubmatch(resp.String())
	if !gjson.Valid(arr[1]) {
		return nil, errors.New("cannot find ytInitialPlayerResponse from html")
	}
	return &Youtube{
		VideoID:    arr[1],
		client:     client,
		jsonResult: gjson.Parse(arr[1]),
	}, nil
}
func (o *Youtube) GetVideoDetails() (YtVideoDetails, error) {
	var result YtVideoDetails
	if !o.jsonResult.Get("videoDetails").Exists() {
		return result, errors.New("cannot find videoDetails")
	}
	err := json.Unmarshal([]byte(o.jsonResult.Get("videoDetails").Raw), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}
func (o *Youtube) GetCaptions() ([]YtCustomCaption, error) {
	var result YtCaptions
	if !o.jsonResult.Get("captions").Exists() {
		return nil, ErrNoYtCaptions
	}
	err := json.Unmarshal([]byte(o.jsonResult.Get("captions").Raw), &result)
	if err != nil {
		return nil, err
	}
	return result.GetCustomCaptions(), nil
}

type YtCustomCaption struct {
	Name         string `json:"name"`
	LanguageCode string `json:"language_code"`
	URL          string `json:"url"`
	IsAsr        bool   `json:"is_asr"`
	IsTranslated bool   `json:"is_translated"`
}

func (c YtCustomCaption) GetTranscript(proxy string) ([]YtTranscriptText, error) {
	url := c.URL
	_, client, err := MakeHTTPClient(proxy, 15*time.Second)
	if err != nil {
		return nil, err
	}
	resp, err := client.R().Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, errors.New("cannot fetch transcript: " + strconv.Itoa(resp.GetStatusCode()) + "; url: " + url)
	}
	var transcript YtTranscript
	err = xml.Unmarshal(resp.Bytes(), &transcript)
	if err != nil {
		return nil, err
	}
	return transcript.Texts, nil
}

type YtCaptions struct {
	PlayerCaptionsTracklistRenderer YtPlayerCaptionsTracklistRenderer `json:"playerCaptionsTracklistRenderer"`
}

func (y YtCaptions) GetCustomCaptions() []YtCustomCaption {
	var captions []YtCustomCaption
	var firstCaption *YtCustomCaption
	for ix, c := range y.PlayerCaptionsTracklistRenderer.CaptionTracks {
		caption := YtCustomCaption{
			Name:         c.Name.SimpleText,
			LanguageCode: c.LanguageCode,
			URL:          c.BaseUrl,
			IsAsr:        c.Kind == "asr",
			IsTranslated: false,
		}
		captions = append(captions, caption)
		if ix == 0 {
			firstCaption = &caption
		}
	}
	if firstCaption == nil || !y.PlayerCaptionsTracklistRenderer.CaptionTracks[0].IsTranslatable {
		return captions
	}
	for _, t := range y.PlayerCaptionsTracklistRenderer.TranslationLanguages {
		captions = append(captions, YtCustomCaption{
			Name:         t.LanguageName.SimpleText,
			LanguageCode: t.LanguageCode,
			URL:          firstCaption.URL + "&tlang=" + t.LanguageCode,
			IsAsr:        false,
			IsTranslated: true,
		})
	}
	return captions
}

type YtPlayerCaptionsTracklistRenderer struct {
	CaptionTracks          []YtCaptionTrack        `json:"captionTracks"`
	AudioTracks            []YtAudioTrack          `json:"audioTracks"`
	TranslationLanguages   []YtTranslationLanguage `json:"translationLanguages"`
	DefaultAudioTrackIndex int                     `json:"defaultAudioTrackIndex"`
}
type YtTranslationLanguage struct {
	LanguageCode string     `json:"languageCode"`
	LanguageName YtLangName `json:"languageName"`
}
type YtAudioTrack struct {
	CaptionTrackIndices []int `json:"captionTrackIndices"`
}
type YtCaptionTrack struct {
	BaseUrl        string     `json:"baseUrl"`
	Name           YtLangName `json:"name"`
	VssId          string     `json:"vssId"`
	LanguageCode   string     `json:"languageCode"`
	Kind           string     `json:"kind"`
	IsTranslatable bool       `json:"isTranslatable"`
	TrackName      string     `json:"trackName"`
}
type YtLangName struct {
	SimpleText string `json:"simpleText"`
}
type YtVideoDetails struct {
	VideoId                string      `json:"videoId"`
	Title                  string      `json:"title"`
	LengthSeconds          string      `json:"lengthSeconds"`
	Keywords               []string    `json:"keywords"`
	ChannelId              string      `json:"channelId"`
	IsOwnerViewing         bool        `json:"isOwnerViewing"`
	ShortDescription       string      `json:"shortDescription"`
	IsCrawlable            bool        `json:"isCrawlable"`
	Thumbnail              YtThumbnail `json:"thumbnail"`
	AllowRatings           bool        `json:"allowRatings"`
	ViewCount              string      `json:"viewCount"`
	Author                 string      `json:"author"`
	IsLowLatencyLiveStream bool        `json:"isLowLatencyLiveStream"`
	IsPrivate              bool        `json:"isPrivate"`
	IsUnpluggedCorpus      bool        `json:"isUnpluggedCorpus"`
	LatencyClass           string      `json:"latencyClass"`
	IsLiveContent          bool        `json:"isLiveContent"`
}
type YtThumbnail struct {
	Thumbnails []YtThumbNailItem `json:"thumbnails"`
}
type YtThumbNailItem struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type YtTranscript struct {
	XMLName xml.Name           `xml:"transcript"`
	Texts   []YtTranscriptText `xml:"text"`
}
type YtTranscriptText struct {
	Start float64 `xml:"start,attr" json:"start"`
	Dur   float64 `xml:"dur,attr" json:"dur"`
	Value string  `xml:",chardata" json:"value"`
}
