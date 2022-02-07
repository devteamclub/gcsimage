package gcsimage

import (
	"bytes"
	"cloud.google.com/go/storage"
	c "context"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
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

func (b *Bucket) getOriginal(ctx c.Context, id string) ([]byte, error) {
	reader, err := b.handle.Object(id).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	buf := new(bytes.Buffer)
	_, errBytes := buf.ReadFrom(reader)
	if errBytes != nil {
		return nil, errBytes
	}

	return buf.Bytes(), nil
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

func (b *Bucket) Get(ctx c.Context, id string, anchor Anchor, width, height int) ([]byte, error) {
	if width <= 0 && height <= 0 {
		return b.getOriginal(ctx, id)
	}

	key := fmt.Sprintf("%s-%d-%d", id, width, height)
	data, err, exist := b.getByKey(ctx, key)
	if exist {
		return data, err
	}

	objHand := b.handle.Object(id)
	attr, err := objHand.Attrs(ctx)
	if err != nil {
		return nil, err
	}

	if attr.ContentType == "image/webp" {
		return b.getOriginal(ctx, id)
	}

	reader, err := objHand.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	original, errImg := imaging.Decode(reader, imaging.AutoOrientation(true))
	if errImg != nil {
		return nil, errImg
	}

	modified := imaging.Fill(original, width, height, imaging.Anchor(anchor), imaging.Lanczos)
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
		return nil, err
	}

	data = buf.Bytes()
	errSave := b.Save(ctx, key, data)
	if errSave != nil {
		return nil, errSave
	}

	return data, nil
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
