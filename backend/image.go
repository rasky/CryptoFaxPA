package main

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/vmihailenco/msgpack"

	"github.com/go-redis/cache"

	dither "github.com/esimov/dithergo"
	resize "github.com/nfnt/resize"
)

var stucki = dither.Dither{
	"Stucki",
	dither.Settings{
		[][]float32{
			[]float32{0.0, 0.0, 0.0, 8.0 / 42.0, 4.0 / 42.0},
			[]float32{2.0 / 42.0, 4.0 / 42.0, 8.0 / 42.0, 4.0 / 42.0, 2.0 / 42.0},
			[]float32{1.0 / 42.0, 2.0 / 42.0, 4.0 / 42.0, 2.0 / 42.0, 1.0 / 42.0},
		},
	},
}

// Load an image (PNG or JPG), resize, convert to monochrome and return as PNG
func ConvertImage(in []byte, width uint) ([]byte, error) {
	orig, _, err := image.Decode(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}

	img := resize.Resize(width, 0, orig, resize.Lanczos3)
	imgmono := stucki.Monochrome(img, 1.0)

	var out bytes.Buffer
	png.Encode(&out, imgmono)
	return out.Bytes(), nil
}

// SaveImagePNG save an image to a PNG file.
func SaveImagePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	png.Encode(f, img)
	return nil
}

type ImageCache struct {
	cache *cache.Codec
}

func NewImageCache(redisUrl string) (*ImageCache, error) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}

	inst := redis.NewClient(opts)
	if cmd := inst.Ping(); cmd.Err() != nil {
		return nil, err
	}

	cache := &cache.Codec{
		Redis: inst,

		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}

	return &ImageCache{cache: cache}, nil
}

func (ic *ImageCache) Set(key string, object interface{}, expiration time.Duration) error {
	return ic.cache.Set(&cache.Item{
		Key:        key,
		Object:     object,
		Expiration: expiration,
	})
}

func (ic *ImageCache) Get(key string, object interface{}) error {
	return ic.cache.Get(key, &object)
}
