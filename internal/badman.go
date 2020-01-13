package internal

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/m-mizutani/badman"
	"github.com/m-mizutani/badman/source"
	"github.com/pkg/errors"
)

func badmanDataPath(prefix string) string {
	return prefix + "badman.msg.gz"
}

func badmanSeralizer() badman.Serializer {
	return badman.NewGzipMsgpackSerializer()
}

// UpdateBadMan updates blacklist data on S3
func UpdateBadMan(s3region, s3bucket, s3prefix string) error {
	s3client := s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String(s3region),
	})))

	man := badman.New()
	man.ReplaceSerializer(badmanSeralizer())
	s3key := badmanDataPath(s3prefix)

	if err := man.Download(source.DefaultSet); err != nil {
		return err
	}

	buf := bytes.Buffer{}
	if err := man.Dump(&buf); err != nil {
		Logger.WithError(err).Errorf("Fail to dump badman repository data: %v", man)
	}

	input := &s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(bytes.NewReader(buf.Bytes())),
		Bucket: &s3bucket,
		Key:    &s3key,
	}

	Logger.WithField("input", input).Trace("Uploading badman repository data")
	resp, err := s3client.PutObject(input)
	if err != nil {
		return errors.Wrapf(err, "Fail to upload badman repository data: %v", input)
	}
	Logger.WithField("output", resp).Trace("Uploaded badman repository data")

	return nil
}

// BuildBadMan creates new badman.BadMan object for matcher
func BuildBadMan(s3region, s3bucket, s3prefix string) (*badman.BadMan, error) {
	s3client := s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String(s3region),
	})))

	input := &s3.GetObjectInput{
		Bucket: &s3bucket,
		Key:    aws.String(badmanDataPath(s3prefix)),
	}

	resp, err := s3client.GetObject(input)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to download filter data: %v", input)
	}

	man := badman.New()
	man.ReplaceSerializer(badmanSeralizer())

	if err := man.Load(resp.Body); err != nil {
		return nil, errors.Wrapf(err, "Fail to load badman data: %v", input)
	}

	return man, nil
}
