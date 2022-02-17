package gcsimage

import (
	"bytes"
	"context"
	"github.com/disintegration/imaging"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
)

var background = context.Background()

func TestInitBucket(t *testing.T) {
	//arrange

	//act
	_, err := InitBucket(background, os.Getenv("IMAGES_STORAGE_BUCKET"))

	//assert
	if err != nil {
		log.Fatalln("fail connect to gcs bucket:", err)
	}
}

func TestGet(t *testing.T) {
	//arrange
	bucket, _ := InitBucket(background, os.Getenv("IMAGES_STORAGE_BUCKET"))

	//act
	goodJPG, ok := bucket.Get(background, "cat", TopRight, 10, 10)
	goodPNG, ok := bucket.Get(background, "cat", TopRight, 10, 10)
	bad, notOk := bucket.Get(background, "", TopRight, 10, 10)

	//assert
	if goodJPG == nil && ok != nil {
		t.Errorf("fail to get jpg image")
	}
	if goodPNG == nil && ok != nil {
		t.Errorf("fail to get png image")
	}

	if bad != nil && notOk == nil {
		t.Errorf("Should error on bad id")
	}
}

func TestGetTransperent(t *testing.T) {
	bucket, err := InitBucket(background, os.Getenv("IMAGES_STORAGE_BUCKET"))
	if err != nil {
		t.Fail()
	}

	original, err := imaging.Open("original.png")
	if err != nil {
		t.Fail()
	}

	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, original, imaging.PNG)
	if err != nil {
		t.Fail()
	}

	id, err := bucket.Add(background, buf.Bytes())
	if err != nil {
		t.Fail()
	}

	data, err := bucket.Get(background, id, Top, 150, 150)
	if err != nil || data == nil {
		t.Fail()
	}
}

func TestResize(t *testing.T) {
	bucket, err := InitBucket(background, os.Getenv("IMAGES_STORAGE_BUCKET"))
	if err != nil {
		t.Fail()
	}

	original, err := imaging.Open("original.png")
	if err != nil {
		t.Fail()
	}

	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, original, imaging.PNG)
	if err != nil {
		t.Fail()
	}

	id, err := bucket.Add(background, buf.Bytes())
	if err != nil {
		t.Fail()
	}

	data, err := bucket.Get(background, id, Top, 150, 0)
	if err != nil || data == nil {
		t.Fail()
	}

	data, err = bucket.Get(background, id, Top, 0, 150)
	if err != nil || data == nil {
		t.Fail()
	}

	data, err = bucket.Get(background, id, Top, 5, 5)
	if err != nil || data == nil {
		t.Fail()
	}
}

func TestOriginal(t *testing.T) {
	bucket, err := InitBucket(background, os.Getenv("IMAGES_STORAGE_BUCKET"))
	if err != nil {
		t.Fail()
	}

	original, err := imaging.Open("original.png")
	if err != nil {
		t.Fail()
	}

	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, original, imaging.PNG)
	if err != nil {
		t.Fail()
	}

	id, err := bucket.Add(background, buf.Bytes())
	if err != nil {
		t.Fail()
	}

	data, err := bucket.Get(background, id, Top, 0, 0)
	if err != nil || data == nil {
		t.Fail()
	}

	data, err = bucket.Get(background, id, Top, -100, 0)
	if err != nil || data == nil {
		t.Fail()
	}

	data, err = bucket.Get(background, id, Top, 0, -100)
	if err != nil || data == nil {
		t.Fail()
	}

	data, err = bucket.Get(background, id, Top, -100, -100)
	if err != nil || data == nil {
		t.Fail()
	}
}

func TestAdd(t *testing.T) {
	//arrange
	bucket, _ := InitBucket(background, os.Getenv("IMAGES_STORAGE_BUCKET"))

	cat := dataFromUrl("https://placekitten.com/500/500")
	empty := make([]byte, 0)

	//act
	err := bucket.Save(background, "cat", cat)
	good, _ := bucket.Add(background, cat)
	bad, _ := bucket.Add(background, empty)

	//assert
	if err != nil {
		t.Errorf("fail to save image")
	}

	if good == "" {
		t.Errorf("fail to add image")
	}

	if bad != "" {
		t.Errorf("Should not add empty image")
	}
}

func TestWebpAdd(t *testing.T) {
	bucket, _ := InitBucket(background, os.Getenv("IMAGES_STORAGE_BUCKET"))

	webpImage, _ := os.ReadFile("5.webp")
	empty := make([]byte, 0)

	//act
	err := bucket.Save(background, "test.webp", webpImage)
	good, _ := bucket.Add(background, webpImage)
	bad, _ := bucket.Add(background, empty)

	//assert
	if err != nil {
		t.Errorf("fail to save image")
	}

	if good == "" {
		t.Errorf("fail to add image")
	}

	if bad != "" {
		t.Errorf("Should not add empty image")
	}
}

func TestWebpGet(t *testing.T) {
	bucket, err := InitBucket(background, os.Getenv("IMAGES_STORAGE_BUCKET"))
	if err != nil {
		t.Errorf("Could not intialize bucket")
	}

	goodWebp, err := bucket.Get(background, "5.webp", Top, 25, 25)
	if err != nil {
		t.Errorf("Could not intialize bucket")
	}

	_, err = bucket.Get(background, "5.webp", Top, 0, 0)
	if err != nil {
		t.Errorf("failed to get existing image with 0 width, height")
	}

	notExistingWebp, notExistingWepbErr := bucket.Get(background, "ThisP1ct9eShou111dNeverBE5.webp", Top, 25, 25)
	if notExistingWepbErr == nil || notExistingWebp != nil {
		t.Fail()
	}

	if len(goodWebp) == 0 || goodWebp == nil {
		t.Errorf("failed to get existing image (check image in the bucket)")
	}

	if notExistingWebp != nil || notExistingWepbErr == nil {
		t.Errorf("Should return nil and error")
	}
}

func dataFromUrl(url string) []byte {
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			return bodyBytes
		}
	}

	return nil
}
