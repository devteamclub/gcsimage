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
	if err != nil {
		t.Fail()
	}

	err = ioutil.WriteFile("resized.png", data, 777)
	if err != nil {
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
