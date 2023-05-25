package delay

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ennismar/go-helper/pkg/constant"
	"github.com/ennismar/go-helper/pkg/log"
	"github.com/ennismar/go-helper/pkg/req"
	"github.com/ennismar/go-helper/pkg/resp"
	"github.com/ennismar/go-helper/pkg/utils"
	"github.com/golang-module/carbon/v2"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Export struct {
	ops   ExportOptions
	Error error
}

func NewExport(options ...func(*ExportOptions)) *Export {
	ops := getExportOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	ex := &Export{
		ops: *ops,
	}
	if ops.dbNoTx == nil {
		log.WithContext(ops.ctx).Warn(ErrDbNil)
		ex.Error = errors.WithStack(ErrDbNil)
	}
	return ex
}

// MigrateExport mysql DDL migrate rollback is not supported, Migrate before New
func MigrateExport(options ...func(*ExportOptions)) (err error) {
	ex := NewExport(options...)
	if ex.Error != nil {
		err = ex.Error
		return
	}
	session := ex.initSession()
	err = session.AutoMigrate(
		new(ExportHistory),
	)
	return
}

func (ex Export) Start(uid, name, category, progress string) (err error) {
	if ex.Error != nil {
		err = ex.Error
		return
	}
	id := strings.TrimSpace(uid)
	if id == "" {
		log.WithContext(ex.ops.ctx).Error(errors.Wrap(ErrUuidNil, id))
		err = errors.WithStack(ErrUuidNil)
		return
	}
	session := ex.initSession().Begin()
	var h ExportHistory
	h.Uuid = id
	h.Category = category
	h.Name = name
	h.Progress = progress
	err = session.
		Model(&ExportHistory{}).
		Create(&h).Error
	if err != nil {
		log.WithContext(ex.ops.ctx).Error(err)
		session.Rollback()
		return
	}
	session.Commit()
	return
}

func (ex Export) Pending(uid, progress string) (err error) {
	if ex.Error != nil {
		err = ex.Error
		return
	}
	id := strings.TrimSpace(uid)
	if id == "" {
		log.WithContext(ex.ops.ctx).Error(errors.Wrap(ErrUuidNil, id))
		err = errors.WithStack(ErrUuidNil)
		return
	}
	session := ex.initSession().Begin()
	var h ExportHistory
	session.
		Model(&ExportHistory{}).
		Where("uuid = ?", id).
		First(&h)
	if h.Id == constant.Zero {
		log.WithContext(ex.ops.ctx).Error(errors.Wrap(ErrUuidInvalid, id))
		err = errors.WithStack(ErrUuidInvalid)
		return
	}
	err = session.
		Model(&ExportHistory{}).
		Where("uuid = ?", uid).
		Update("progress", progress).Error
	if err != nil {
		log.WithContext(ex.ops.ctx).Error(err)
		session.Rollback()
		return
	}
	session.Commit()
	return
}

func (ex Export) End(uid string, args ...string) (err error) {
	if ex.Error != nil {
		err = ex.Error
		return
	}
	id := strings.TrimSpace(uid)
	if id == "" {
		log.WithContext(ex.ops.ctx).Error(errors.Wrap(ErrUuidNil, id))
		err = errors.WithStack(ErrUuidNil)
		return
	}
	session := ex.initSession().Begin()
	var h ExportHistory
	session.
		Model(&ExportHistory{}).
		Where("uuid = ?", id).
		First(&h)
	if h.Id == constant.Zero {
		log.WithContext(ex.ops.ctx).Error(errors.Wrap(ErrUuidInvalid, id))
		err = errors.WithStack(ErrUuidInvalid)
		return
	}

	m := make(map[string]interface{})
	var progress, filename string
	switch len(args) {
	case 1:
		progress = args[0]
	case 2:
		progress = args[0]
		filename = args[1]
	}
	if filename != "" {
		// 文件名不为空需要上传文件到oss
		//var bucket *oss.Bucket
		//bucket, err = ex.getBucket()
		//if err != nil {
		//	err = errors.WithStack(err)
		//	return
		//}
		s3Api := ex.getS3Client()

		objName := fmt.Sprintf("%s/%s/%s/%s", ex.ops.objPrefix, carbon.Now().ToDateString(), ex.ops.machineId, filepath.Base(filename))
		//err = bucket.PutObjectFromFile(objName, filename)
		err = ex.PutObject(s3Api, objName, filename)
		if err != nil {
			log.WithContext(ex.ops.ctx).Error(errors.Wrap(ErrOssPutObjectFailed, err.Error()))
			err = errors.WithStack(ErrOssPutObjectFailed)
			return
		}
		m["url"] = objName
	}
	m["progress"] = progress
	m["end"] = constant.One
	err = session.
		Model(&ExportHistory{}).
		Where("uuid = ?", uid).
		Updates(&m).Error
	if err != nil {
		session.Rollback()
		log.WithContext(ex.ops.ctx).Error(err)
		return
	}
	session.Commit()
	return
}

// FindHistory query export history list
func (ex Export) FindHistory(r *req.DelayExportHistory) (rp []resp.DelayExportHistory, err error) {
	if ex.Error != nil {
		err = ex.Error
		return
	}
	session := ex.initSession()
	list := make([]ExportHistory, 0)
	q := session.
		Model(&ExportHistory{}).
		Order("created_at DESC")
	name := strings.TrimSpace(r.Name)
	if name != "" {
		q.Where("name LIKE ?", fmt.Sprintf("%%%s%%", name))
	}
	category := strings.TrimSpace(r.Category)
	if category != "" {
		q.Where("category LIKE ?", fmt.Sprintf("%%%s%%", category))
	}
	if r.End != nil {
		q.Where("end = ?", *r.End)
	}
	page := &r.Page
	countCache := false
	if page.CountCache != nil {
		countCache = *page.CountCache
	}
	if !page.NoPagination {
		if !page.SkipCount {
			q.Count(&page.Total)
		}
		if page.Total > 0 || page.SkipCount {
			limit, offset := page.GetLimit()
			q.Limit(limit).Offset(offset).Find(&list)
		}
	} else {
		// no pagination
		q.Find(&list)
		page.Total = int64(len(list))
		page.GetLimit()
	}
	page.CountCache = &countCache
	//var bucket *oss.Bucket
	//bucket, err = ex.getBucket()
	//if err != nil {
	//	err = errors.WithStack(err)
	//	return
	//}
	s3cli := ex.getS3Client()
	for i, item := range list {
		if item.End == constant.One && item.Url != "" {
			// get signature url
			var url string
			s3rsp, _ := s3cli.GetObjectRequest(&s3.GetObjectInput{
				Bucket: aws.String(ex.ops.bucket),
				Key:    aws.String(item.Url),
			})
			url, err := s3rsp.Presign(time.Duration(ex.ops.expire) * time.Minute)
			//url, err = bucket.SignURL(item.Url, http.MethodGet, ex.ops.expire*60)
			if err != nil {
				continue
			}
			list[i].Url = url
		}
	}
	rp = make([]resp.DelayExportHistory, 0)
	utils.Struct2StructByJson(list, &rp)
	return
}

// DeleteHistoryByIds delete export history
func (ex Export) DeleteHistoryByIds(ids []uint) (err error) {
	if ex.Error != nil {
		err = ex.Error
		return
	}
	session := ex.initSession().Begin()
	list := make([]ExportHistory, 0)
	session.
		Model(&ExportHistory{}).
		Where("id IN (?)", ids).
		Find(&list)
	//endObjs := make([]string, 0)
	//for _, item := range list {
	//	if item.End == constant.One {
	//		endObjs = append(endObjs, item.Url)
	//	}
	//}
	//if len(endObjs) > 0 {
	//	var bucket *oss.Bucket
	//	bucket, err = ex.getBucket()
	//	if err != nil {
	//		session.Rollback()
	//		err = errors.WithStack(err)
	//		return
	//	}
	//	_, err = bucket.DeleteObjects(endObjs)
	//	if err != nil {
	//		session.Rollback()
	//		err = errors.WithStack(err)
	//		return
	//	}
	//}
	objs := make([]*s3.ObjectIdentifier, 0)
	for _, v := range list {
		if v.End == constant.One {
			objs = append(objs, &s3.ObjectIdentifier{
				Key: aws.String(v.Url),
			})
		}
	}
	if len(objs) > 0 {
		s3cli := ex.getS3Client()
		deleteObjectsInput := &s3.DeleteObjectsInput{
			Bucket: aws.String(ex.ops.bucket),
			Delete: &s3.Delete{
				Objects: objs,
				Quiet:   aws.Bool(false),
			},
		}
		_, err = s3cli.DeleteObjects(deleteObjectsInput)
		if err != nil {
			session.Rollback()
			err = errors.WithStack(err)
			return
		}
	}
	err = session.
		Where("id IN (?)", ids).
		Delete(&ExportHistory{}).Error
	if err != nil {
		session.Rollback()
		return
	}
	session.Commit()
	return
}

func (ex Export) getS3Client() *s3.S3 {
	region := "cn"
	config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(ex.ops.key, ex.ops.secret, ""),
		Endpoint:         aws.String(ex.ops.endpoint),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(true),
		LogLevel:         aws.LogLevel(aws.LogDebug),
		Region:           aws.String(region),
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost: 1000,
			},
		},
	}
	sess := session.Must(session.NewSession(config))

	return s3.New(sess)
}
func (ex Export) PutObject(s3cli *s3.S3, key string, fn string) error {

	file, err := os.Open(fn)
	if err != nil {
		log.WithContext(ex.ops.ctx).Error(errors.Wrap(ErrOssPutObjectFailed, err.Error()))
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.WithContext(ex.ops.ctx).Error(errors.Wrap(ErrOssPutObjectFailed, err.Error()))
		}
	}()
	_, err = s3cli.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(ex.ops.bucket),
		Key:    aws.String(key),
		Body:   file,
		//StorageClass: aws.String(s3.StorageClassGlacier),
	})

	if err != nil {
		log.WithContext(ex.ops.ctx).Error(errors.Wrap(ErrOssPutObjectFailed, err.Error()))
		return err
	}
	return nil
}

//func (ex Export) getBucket() (bucket *oss.Bucket, err error) {
//	var client *oss.Client
//	client, err = oss.New(ex.ops.endpoint, ex.ops.key, ex.ops.secret)
//	if err != nil {
//		log.WithContext(ex.ops.ctx).Error(errors.Wrap(ErrOssSecretInvalid, err.Error()))
//		err = errors.WithStack(ErrOssSecretInvalid)
//		return
//	}
//	bucket, err = client.Bucket(ex.ops.bucket)
//	if err != nil {
//		log.WithContext(ex.ops.ctx).Error(errors.Wrap(ErrOssBucketInvalid, err.Error()))
//		err = errors.WithStack(ErrOssBucketInvalid)
//		return
//	}
//	return
//}

func (ex Export) initSession() *gorm.DB {
	namingStrategy := schema.NamingStrategy{
		TablePrefix:   ex.ops.tbPrefix,
		SingularTable: true,
	}
	session := ex.ops.dbNoTx.WithContext(ex.ops.ctx).Session(&gorm.Session{})
	session.NamingStrategy = namingStrategy
	return session
}
