package main

import (
	"aws-iam/api/handler"
	"aws-iam/entity"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

type RoleCreds struct {
	accessKey    string
	secretKey    string
	sessionToken string
}

func main() {
	router := handler.Init()
	router.Run("localhost:8080")
	//sandbox()
}

func sandbox() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Println(err)
	}

	creds := assumeRoleAndGenerateCreds(ROLE_READ, cfg)
	listS3(BUCKET_NAME, cfg, creds)

	creds = assumeRoleAndGenerateCreds(ROLE_WRITE, cfg)
	getObject(BUCKET_NAME, "key-yakir-343-246-234-324-329", cfg, creds)

	creds = assumeRoleAndGenerateCreds(ROLE_WRITE, cfg)
	putObject(BUCKET_NAME, cfg, creds, &entity.MyObject{
		Key:     "key-yakir-343-246-234-324-329",
		Content: "this is a test body, bla bla",
	})

	creds = assumeRoleAndGenerateCreds(ROLE_WRITE, cfg)
	getObject(BUCKET_NAME, "key-yakir-343-246-234-324-329", cfg, creds)

}

func assumeRoleAndGenerateCreds(roleArn string, cfg aws.Config) *RoleCreds {
	role, err := AssumeRoleARN(roleArn, cfg)
	if err != nil {
		log.Println("Failed to assume role", ROLE_READ)
		return nil
	}

	creds := generateCreds(role)

	return creds
}

func getObject(bucketName string, key string, cfg aws.Config, creds *RoleCreds) {
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(creds.accessKey, creds.secretKey, creds.sessionToken))
	})

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	resp, err := client.GetObject(context.TODO(), input)
	if err != nil {
		log.Println("Failed to find object ", key, " in bucket ", bucketName)
		log.Println(err)
		return
	}

	defer resp.Body.Close()
	buffer := make([]byte, resp.ContentLength)
	resp.Body.Read(buffer)
	log.Println(string(buffer))
}

func generateCreds(role *sts.AssumeRoleOutput) *RoleCreds {
	return &RoleCreds{
		accessKey:    *(role.Credentials.AccessKeyId),
		secretKey:    *(role.Credentials.SecretAccessKey),
		sessionToken: *(role.Credentials.SessionToken),
	}
}

func AssumeRoleARN(roleArn string, cfg aws.Config) (*sts.AssumeRoleOutput, error) {
	stsClient := sts.NewFromConfig(cfg)

	role, err := stsClient.AssumeRole(context.TODO(), &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("yakir-iam-project"),
		ExternalId:      aws.String(EXTERNAL_ID),
	})

	if err == nil {
		log.Println("Role", roleArn, "assumed successfully")
		return role, nil
	}

	return nil, err
}

func putObject(bucketName string, cfg aws.Config, creds *RoleCreds, object *entity.MyObject) (err error) {
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(creds.accessKey, creds.secretKey, creds.sessionToken))
	})

	log.Println(object.Content)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(object.Key),
		Body:   strings.NewReader(object.Content),
	}

	x, err := client.PutObject(context.TODO(), input)
	if err != nil {
		log.Println("Failed to upload", object)
		return
	}

	log.Println(x)
	return nil
}

func listS3(bucketName string, cfg aws.Config, creds *RoleCreds) (err error) {
	//client := s3.New(s3.Options{
	//	Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken)),
	//	Region:      cfg.Region,
	//})
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(creds.accessKey, creds.secretKey, creds.sessionToken))
	})

	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		log.Println(err)
		return
	}

	for _, object := range output.Contents {
		log.Printf("key=%s size=%d", aws.ToString(object.Key), object.Size)
	}

	return nil
}
