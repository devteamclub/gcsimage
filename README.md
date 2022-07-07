# Google Cloud Storage images
This wrapper is based on [github.com/disintegration/imaging](https://github.com/disintegration/imaging). Big thanks to @disintegration


## Setup
```
GOOGLE_APPLICATION_CREDENTIALS=gcs_key.json
IMAGES_STORAGE_BUCKET=anthive-img
```


## How to use it
```Go
import 	"github.com/devteamclub/gcsimage"

var bucket *gcsimage.Bucket

func main() {
    initImageBucket(context.Background())
	
    // ... //
    id, err := bucket.Add(context.Background(), data)
	
    // ... //
    data, contentType, err := bucket.Get(context.Background(), id, gcsimage.Top, width, height)
}

func initImageBucket(ctx context.Context) {
    var err error
    bucket, err = gcsimage.InitBucket(ctx, os.Getenv("IMAGES_STORAGE_BUCKET"))
    if err != nil {
    panic(err)
    }
}

```
