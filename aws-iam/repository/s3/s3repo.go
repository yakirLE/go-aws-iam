package s3repository

import (
	"aws-iam/entity"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"log"
	"strings"
)

const (
	ROLE_READ   = "arn:aws:iam::983604318039:role/Cross_account_read_bucket_iam_project"
	ROLE_WRITE  = "arn:aws:iam::983604318039:role/Cross_account_limited_action_iam_project"
	EXTERNAL_ID = "axiom"
	BUCKET_NAME = "iam-project"
)

var cfg aws.Config

func PutObject(object *entity.MyObject) error {
	client, err := generateS3Client(ROLE_WRITE)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(object.Key),
		Body:   strings.NewReader(object.Content),
	}

	_, err = client.PutObject(context.TODO(), input)
	if err != nil {
		log.Println("Failed to upload", object)
		return err
	}

	return nil
}

func GetObject(key string) (*entity.MyObject, error) {
	client, err := generateS3Client(ROLE_WRITE)
	if err != nil {
		return nil, err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(key),
	}

	resp, err := client.GetObject(context.TODO(), input)
	if err != nil {
		log.Println("Failed to find object ", key, " in bucket ", BUCKET_NAME)
		return nil, err
	}

	defer resp.Body.Close()
	buffer := make([]byte, resp.ContentLength)
	resp.Body.Read(buffer)

	return &entity.MyObject{
		Key:     key,
		Content: string(buffer),
	}, nil
}

func generateS3Client(roleArn string) (*s3.Client, error) {
	role, err := assumeRoleARN(roleArn)
	if err != nil {
		return nil, err
	}

	client, err := generateConnection(role)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func generateConnection(role *sts.AssumeRoleOutput) (client *s3.Client, err error) {
	cfg, err = config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Println(err)
		return
	}

	if role != nil {
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(*role.Credentials.AccessKeyId, *role.Credentials.SecretAccessKey, *role.Credentials.SessionToken))
		})
		//client := s3.New(s3.Options{
		//	Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken)),
		//	Region:      cfg.Region,
		//})
	}

	return
}

func assumeRoleARN(roleArn string) (*sts.AssumeRoleOutput, error) {
	_, err := generateConnection(nil)
	if err != nil {
		return nil, err
	}

	stsClient := sts.NewFromConfig(cfg)

	role, err := stsClient.AssumeRole(context.TODO(), &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("yakir-iam-project"),
		ExternalId:      aws.String(EXTERNAL_ID),
	})

	if err != nil {
		return nil, err
	}

	log.Println("Role", roleArn, "assumed successfully")
	return role, nil
}

func convertS3ObjectToMyObject(object types.Object, content string) entity.MyObject {
	return entity.MyObject{
		Key:     *(object.Key),
		Content: content,
	}
}

func ListS3Bucket() ([]string, error) {
	client, err := generateS3Client(ROLE_READ)
	if err != nil {
		return nil, err
	}

	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(BUCKET_NAME),
	})

	if err != nil {
		log.Println(err)
		return nil, err
	}

	var keys []string
	for _, object := range output.Contents {
		keys = append(keys, *object.Key)
	}

	return keys, nil
}

func ListAndDownloadS3Bucket() (*[]entity.MyObject, error) {
	keys, err := ListS3Bucket()
	if err != nil {
		return nil, err
	}

	var myObjects []entity.MyObject
	for _, key := range keys {
		object, err := GetObject(key)
		if err != nil {
			log.Println("Failed to download", key, ", error:", err)
			continue
		}

		myObjects = append(myObjects, *object)
	}

	return &myObjects, nil
}

/*
func downloadObject(key string) (string, error) {
	output, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(output.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
*/
