package db

import (
	"fmt"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/pkg/errors"
)

// --- Certificate Functions ---

func GetCertificates(pageIndex, pageSize int) (certs []model.Certificate, count int64, err error) {
	certDB := db.Model(&model.Certificate{})
	if err := certDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get certificates count")
	}
	if err := certDB.Order(fmt.Sprintf("%s DESC", columnName("id"))).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&certs).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed find certificates")
	}
	return certs, count, nil
}

func GetCertificateByID(id uint) (*model.Certificate, error) {
	var cert model.Certificate
	if err := db.First(&cert, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get certificate by id: %d", id)
	}
	return &cert, nil
}

// GetCertificateByOwnerID 根据所有者ID获取证书，这是租户端查询的核心
func GetCertificateByOwnerID(ownerID uint) (*model.Certificate, error) {
	var cert model.Certificate
	// 一个租户只应该有一个有效证书，所以使用 First
	// 只查询状态为 valid 或 expiring 且未过期的证书
	if err := db.Where("owner_id = ? AND (status = ? OR status = ?) AND expiration_date > ?",
		ownerID, model.CertificateStatusValid, model.CertificateStatusExpiring, time.Now()).First(&cert).Error; err != nil {
		return nil, err // GORM 会在找不到记录时返回 ErrRecordNotFound
	}
	return &cert, nil
}

func CreateCertificate(cert *model.Certificate) error {
	return errors.WithStack(db.Create(cert).Error)
}

func UpdateCertificate(cert *model.Certificate) error {
	return errors.WithStack(db.Save(cert).Error)
}

func DeleteCertificate(id uint) error {
	return errors.WithStack(db.Delete(&model.Certificate{}, id).Error)
}

// --- CertificateRequest Functions ---

func GetCertificateRequests(pageIndex, pageSize int) (reqs []model.CertificateRequest, count int64, err error) {
	reqDB := db.Model(&model.CertificateRequest{})
	if err := reqDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get certificate requests count")
	}
	if err := reqDB.Order(fmt.Sprintf("%s DESC", columnName("id"))).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&reqs).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed find certificate requests")
	}
	return reqs, count, nil
}

func GetCertificateRequestByID(id uint) (*model.CertificateRequest, error) {
	var req model.CertificateRequest
	if err := db.First(&req, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get certificate request by id: %d", id)
	}
	return &req, nil
}

// GetCertificateRequestsByUserID 获取某个用户的所有申请记录
func GetCertificateRequestsByUserID(userID uint) ([]model.CertificateRequest, error) {
	var requests []model.CertificateRequest
	if err := db.Where("user_id = ?", userID).Order(fmt.Sprintf("%s DESC", columnName("id"))).Find(&requests).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get certificate requests for user id: %d", userID)
	}
	return requests, nil
}

// GetPendingCertificateRequestByUserID 检查用户是否已有待处理的申请
func GetPendingCertificateRequestByUserID(userID uint) (*model.CertificateRequest, error) {
	var request model.CertificateRequest
	if err := db.Where("user_id = ? AND status = ?", userID, model.CertificateStatusPending).First(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func CreateCertificateRequest(req *model.CertificateRequest) error {
	return errors.WithStack(db.Create(req).Error)
}

func UpdateCertificateRequest(req *model.CertificateRequest) error {
	return errors.WithStack(db.Save(req).Error)
}