package storage

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// R2Config はCloudflare R2の設定を表します
type R2Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
}

// GetR2Client はCloudflare R2クライアントを作成します
func GetR2Client(config R2Config) (*s3.S3, error) {
	// AWS設定
	awsConfig := &aws.Config{
		Endpoint:         aws.String(config.Endpoint),
		Region:           aws.String("auto"), // Cloudflare R2ではregionは使用されない
		Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
		S3ForcePathStyle: aws.Bool(true), // パススタイルのURLを使用
	}

	// セッションを作成
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// S3クライアントを作成
	return s3.New(sess), nil
}

// UploadImageToR2 は画像をCloudflare R2にアップロードします
func UploadImageToR2(client *s3.S3, bucketName, localFilePath, objectKey string) error {
	// ファイルを開く
	file, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// アップロードパラメータを設定
	params := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
		ACL:    aws.String("public-read"), // 必要に応じて変更
	}

	// ファイルをアップロード
	_, err = client.PutObject(params)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}

// GetR2ObjectURL はCloudflare R2オブジェクトのURLを生成します
func GetR2ObjectURL(endpoint, bucketName, objectKey string) string {
	// エンドポイントの末尾のスラッシュを削除
	if endpoint[len(endpoint)-1] == '/' {
		endpoint = endpoint[:len(endpoint)-1]
	}

	// オブジェクトキーの先頭のスラッシュを削除
	if len(objectKey) > 0 && objectKey[0] == '/' {
		objectKey = objectKey[1:]
	}

	// URLを生成
	return fmt.Sprintf("%s/%s/%s", endpoint, bucketName, objectKey)
}
