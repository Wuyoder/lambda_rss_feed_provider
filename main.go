package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Feed represents feed.json structure
type Feed struct {
	Podcast struct {
		LastBuildDate string `json:"lastBuildDate"`
		Title         string `json:"title"`
		Description   string `json:"description"`
		Image         struct {
			URL   string `json:"url"`
			Title string `json:"title"`
			Link  string `json:"link"`
		} `json:"image"`
		Link         string `json:"link"`
		Author       string `json:"author"`
		Copyright    string `json:"copyright"`
		Email        string `json:"email"`
		Language     string `json:"language"`
		Type         string `json:"type"`
		CategoryMain string `json:"category_main"`
		CategorySub  string `json:"category_sub"`
		Explicit     string `json:"explicit"`
		Episodes     []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Link        string `json:"link"`
			Audio       string `json:"audio"`
			Creator     string `json:"creator"`
			Explicit    string `json:"explicit"`
			Duration    string `json:"duration"`
			Image       string `json:"image"`
			Episode     int    `json:"episode"`
			Type        string `json:"type"`
		} `json:"episodes"`
	} `json:"podcast"`
}

// GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap -tags lambda.norpc main.go
// zip myFunction.zip bootstrap

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, s3Event events.S3Event) error {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess)
	object := s3Event.Records[0].S3.Object
	bucket := "intoxicating"
	key := object.Key

	// Download object from S3
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		fmt.Println("Error while downloading the object", err)
		return err
	}

	defer resp.Body.Close()

	// Decode JSON payload into struct
	var feed Feed
	json.NewDecoder(resp.Body).Decode(&feed)

	// Generate RSS feed based on feed.json
	fmt.Println(feed)
	rssFeed := generateRSS(feed)
	fmt.Println("RSS feed input:")

	// Upload RSS feed to S3 root path as feed.rss
	var body bytes.Buffer
	body.Write(rssFeed)
	fmt.Println("RSS feed output:")
	fmt.Println(string(rssFeed))
	uploadResp, err := svc.PutObject(&s3.PutObjectInput{
		Body:        bytes.NewReader(body.Bytes()),
		Bucket:      aws.String(bucket),
		Key:         aws.String("feed.rss"),
		ContentType: aws.String("application/xml"),
	})
	if err != nil {
		fmt.Println("Error while uploading the object", err)
		return err
	}

	fmt.Printf("Feed uploaded to S3 with status %v\n", uploadResp)

	return nil
}

func generateRSS(feed Feed) []byte {
	rss := `<?xml version="1.0" encoding="UTF-8"?><rss xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:atom="http://www.w3.org/2005/Atom" version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">`
	rss += `<channel>`
	rss += fmt.Sprintf(`<title><![CDATA[%s]]></title>`, feed.Podcast.Title)
	rss += fmt.Sprintf(`<description><![CDATA[%s]]></description>`, feed.Podcast.Description)
	rss += fmt.Sprintf(`<link>%s</link>`, feed.Podcast.Link)
	rss += fmt.Sprintf(`<image><url>%s</url><title>%s</title><link>%s</link></image>`, feed.Podcast.Image.URL, feed.Podcast.Image.Title, feed.Podcast.Image.Link)
	rss += fmt.Sprintf(`<generator>%s</generator>`, "yodering.net")
	rss += fmt.Sprintf(`<lastBuildData>%s</lastBuildData>`, feed.Podcast.LastBuildDate)
	rss += fmt.Sprint(`<atom:link href="https://yodering.net/feed.rss" rel="self" type="application/rss+xml"/>`)
	rss += fmt.Sprintf(`<author><![CDATA[%s]]></author>`, feed.Podcast.Author)
	rss += fmt.Sprintf(`<copyright><![CDATA[%s]]></copyright>`, feed.Podcast.Author)
	rss += fmt.Sprintf("<language>%s</language>", feed.Podcast.Language)
	rss += fmt.Sprintf(`<category><![CDATA[%s]]></category>`, feed.Podcast.CategoryMain)
	rss += fmt.Sprintf(`<category><![CDATA[%s]]></category>`, feed.Podcast.CategorySub)
	rss += fmt.Sprintf(`<itunes:author>%s</itunes:author>`, feed.Podcast.Author)
	rss += fmt.Sprintf(`<itunes:summary><![CDATA[%s]]></itunes:summary>`, feed.Podcast.Description)
	rss += fmt.Sprintf(`<itunes:image href="%s"/>`, feed.Podcast.Image.URL)
	rss += fmt.Sprintf(`<itunes:type>%s</itunes:type>`, feed.Podcast.Type)
	rss += fmt.Sprint(`<itunes:owner>`)
	rss += fmt.Sprintf(`<itunes:name><![CDATA[%s]]></itunes:name>`, feed.Podcast.Author)
	rss += fmt.Sprintf(`<itunes:email>%s</itunes:email>`, feed.Podcast.Email)
	rss += fmt.Sprintf(`</itunes:owner>`)
	rss += fmt.Sprintf(`<itunes:explicit>%s</itunes:explicit>`, feed.Podcast.Explicit)
	rss += fmt.Sprintf(`<itunes:category text="%s"> <itunes:category text="%s" /></itunes:category>`, feed.Podcast.CategoryMain, feed.Podcast.CategorySub)
	for _, v := range feed.Podcast.Episodes {
		rss += fmt.Sprintf(`<item>`)
		rss += fmt.Sprintf(`<title><![CDATA[%s]]></title>`, v.Title)
		rss += fmt.Sprintf(`<description><![CDATA[%s]]></description>`, v.Description)
		rss += fmt.Sprintf(`<link>%s</link>`, v.Link)
		rss += fmt.Sprintf(`<guid>%s</guid>`, strings.Split(v.Link, "/")[:len(strings.Split(v.Link, "/"))-1])
		rss += fmt.Sprintf(`<dc:creator><![CDATA[%s]]></dc:creator>`, v.Creator)
		rss += fmt.Sprintf(`<pubDate>%s</pubDate>`, feed.Podcast.LastBuildDate)
		rss += fmt.Sprintf(`<enclosure url="%s" length="%s" type="audio/mpeg"/>`, v.Audio, v.Duration)
		rss += fmt.Sprintf(`<itunes:summary><![CDATA[%s]]></itunes:summary>`, v.Description)
		rss += fmt.Sprintf(`<itunes:explicit>%s</itunes:explicit>`, v.Explicit)
		rss += fmt.Sprintf(`<itunes:duration>%s</itunes:duration>`, v.Duration)
		rss += fmt.Sprintf(`<itunes:image href="%s"/>`, v.Image)
		rss += fmt.Sprintf(`<itunes:episode>%d</itunes:episode>`, v.Episode)
		rss += fmt.Sprintf(`<itunes:episodeType>%s</itunes:episodeType>`, v.Type)
		rss += fmt.Sprintf(`</item>`)
	}
	rss += `</channel></rss>`
	return []byte(rss)
}
