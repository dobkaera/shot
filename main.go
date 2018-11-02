package main

import (
	"bytes"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/kbinani/screenshot"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/go-ini/ini.v1"
	"image"
	"image/png"
	"os"
	"path"
)

type Announcer struct {
	config *ini.File
}

func (a *Announcer) LoadConfig() error {
	dir, err := homedir.Dir()

	if nil != err {
		return err
	}

	configPath := path.Join(dir, "Dropbox", ".shot.ini")

	config, err := ini.Load(configPath)

	if nil != err {
		return err
	}

	a.config = config

	return nil
}

func (a *Announcer) getImages() ([]*image.RGBA, error) {
	n := screenshot.NumActiveDisplays()

	result := make([]*image.RGBA, 0)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)

		img, err := screenshot.CaptureRect(bounds)

		if err != nil {
			return result, nil
		}

		result = append(result, img)
	}

	return result, nil
}

func (a *Announcer) Publish(text string) error {
	images, err := a.getImages()

	if nil != err {
		return err
	}

	consumerKey, err := a.config.Section("twitter").GetKey("consumer_key")

	if nil != err {
		return err
	}

	consumerSecret, err := a.config.Section("twitter").GetKey("consumer_secret")

	if nil != err {
		return err
	}

	accessToken, err := a.config.Section("twitter").GetKey("access_token")

	if nil != err {
		return err
	}

	accessSecret, err := a.config.Section("twitter").GetKey("access_secret")

	if nil != err {
		return err
	}

	config := oauth1.NewConfig(consumerKey.Value(), consumerSecret.Value())
	token := oauth1.NewToken(accessToken.Value(), accessSecret.Value())
	// OAuth1 http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	outImage := &bytes.Buffer{}

	foundIndex := -1

	for i := 0; i < len(images); i++ {
		bounds := images[i].Bounds()

		deltaX := (bounds.Max.X - bounds.Min.X)
		deltaY := (bounds.Max.Y - bounds.Min.Y)

		//fmt.Printf("bounds: %+v deltaX: %d deltaY: %d\n", bounds, deltaX, deltaY)

		// TODO: Make it parameters on the ~/Dropbox/shot.ini file

		if deltaX == 1920 && deltaY == 1080 {
			foundIndex = i
			break
		}
	}

	err = png.Encode(outImage, images[foundIndex])

	if nil != err {
		return err
	}

	tempFile, _ := os.OpenFile("/tmp/image-to-upload.png", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	tempFile.Write(outImage.Bytes())

	tempFile.Close()

	mediaObject, _, err := client.Media.UploadFile("/tmp/image-to-upload.png")

	if nil != err {
		return err
	}

	t, r, err := client.Statuses.Update(text, &twitter.StatusUpdateParams{
		MediaIds: []int64{mediaObject.MediaID},
	})

	fmt.Printf("%+v %+v\r\n", t, r)
	return err
}

func main() {
	a := Announcer{}

	a.LoadConfig()

	a.Publish(os.Args[1])

}
