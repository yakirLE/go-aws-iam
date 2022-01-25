package service

import (
	"aws-iam/entity"
	s3repository "aws-iam/repository/s3"
	"github.com/google/uuid"
)

var cache map[string]*entity.MyObject

func PushObject(object *entity.MyObject) error {
	initCache()
	if object.Key == "" {
		object.Key = uuid.New().String()
	}

	err := s3repository.PutObject(object)
	if err != nil {
		return err
	}

	cache[object.Key] = object
	return nil
}

func GetObject(key string) (*entity.MyObject, error) {
	initCache()
	_, ok := cache[key]
	if !ok {
		object, err := s3repository.GetObject(key)
		if err != nil {
			return nil, err
		}

		cache[key] = object
	}

	return cache[key], nil
}

func ListObjects(fromCache bool) (*[]entity.MyObject, error) {
	if fromCache {
		objects := make([]entity.MyObject, 0, len(cache))
		for _, object := range cache {
			objects = append(objects, *object)
		}

		return &objects, nil
	} else {
		objects, err := s3repository.ListAndDownloadS3Bucket()
		if err != nil {
			return nil, err
		}

		for _, object := range *objects {
			cache[object.Key] = &object
		}

		return objects, nil
	}
}

func initCache() {
	if cache == nil {
		cache = make(map[string]*entity.MyObject)
	}
}
