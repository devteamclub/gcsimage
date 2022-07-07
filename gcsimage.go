package gcsimage

import (
	"bytes"
	"cloud.google.com/go/storage"
	c "context"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"image"
	"io/ioutil"
)

type Anchor int

const (
	Center Anchor = iota
	TopLeft
	Top
	TopRight
	Left
	Right
	BottomLeft
	Bottom
	BottomRight
)

type Bucket struct {
	handle *storage.BucketHandle
}

func InitBucket(ctx c.Context, bucketName string) (*Bucket, error) {
	if len(bucketName) == 0 {
		return nil, errors.New("bucket name is empty")
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Bucket{
		handle: client.Bucket(bucketName),
	}, nil
}

func (b *Bucket) getByKey(ctx c.Context, key string) ([]byte, error, bool) {
	reader, err := b.handle.Object(key).NewReader(ctx)
	if err != nil {
		return nil, err, false
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err, true
	}

	return data, nil, true
}

func (b *Bucket) Get(ctx c.Context, id string, anchor Anchor, width, height int) (data []byte, contentType string, err error) {
	objHand := b.handle.Object(id)
	attr, attrErr := objHand.Attrs(ctx)
	if attrErr != nil {
		err = attrErr
		return
	}

	contentType = attr.ContentType
	if contentType == "image/webp" || (width <= 0 && height <= 0) {
		data, err, _ = b.getByKey(ctx, id)
		return
	}

	key := fmt.Sprintf("%s-%d-%d", id, width, height)
	isExist := false
	data, err, isExist = b.getByKey(ctx, key)
	if isExist {
		return
	}

	reader, readErr := objHand.NewReader(ctx)
	if readErr != nil {
		err = readErr
		return
	}
	defer reader.Close()

	original, imgErr := imaging.Decode(reader, imaging.AutoOrientation(true))
	if imgErr != nil {
		err = imgErr
		return
	}

	var modified *image.NRGBA
	if width > 0 && height > 0 {
		modified = imaging.Fill(original, width, height, imaging.Anchor(anchor), imaging.Lanczos)
	} else {
		modified = imaging.Resize(original, width, height, imaging.Lanczos)
	}

	buf := new(bytes.Buffer)
	switch attr.ContentType {
	case "image/png":
		err = imaging.Encode(buf, modified, imaging.PNG)
	case "image/jpeg":
		err = imaging.Encode(buf, modified, imaging.JPEG)
	case "image/gif":
		err = imaging.Encode(buf, modified, imaging.GIF)
	default:
		msg := fmt.Sprintf("%s is not supported. Only image/png, image/jpeg, image/gif", attr.ContentType)
		err = errors.New(msg)
	}
	if err != nil {
		return
	}

	data = buf.Bytes()
	saveErr := b.Save(ctx, key, data)
	if saveErr != nil {
		err = saveErr
		return
	}

	return
}

func (b *Bucket) Add(ctx c.Context, data []byte) (string, error) {
	id := uuid.New().String()
	err := b.Save(ctx, id, data)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (b *Bucket) Save(ctx c.Context, key string, data []byte) error {
	if len(data) == 0 {
		return errors.New("data is empty")
	}

	writer := b.handle.Object(key).NewWriter(ctx)
	defer writer.Close()
	_, errWrite := writer.Write(data)
	if errWrite != nil {
		return errWrite
	}

	return nil
}
