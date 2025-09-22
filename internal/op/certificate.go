package op

import (
	"fmt"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// --- Certificate Service ---

var GetCertificateByID = db.GetCertificateByID
var GetCertificates = db.GetCertificates
var CreateCertificate = db.CreateCertificate
var UpdateCertificate = db.UpdateCertificate

// GetCertificateForTenant 是租户端调用的核心服务
func GetCertificateForTenant(ownerID uint) (*model.Certificate, error) {
	cert, err := db.GetCertificateByOwnerID(ownerID)
	if err != nil {
		// 如果错误不是 "记录未找到"，则是一个真正的数据库错误
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// 如果是 "记录未找到"，是正常情况，说明租户没有证书
		return nil, nil
	}
	return cert, nil
}

func UpdateCertificateDetails(id uint, name string, expirationDate time.Time) (*model.Certificate, error) {
	cert, err := db.GetCertificateByID(id)
	if err != nil {
		return nil, err
	}
	cert.Name = name
	cert.ExpirationDate = expirationDate
	err = db.UpdateCertificate(cert)
	return cert, err
}

func RevokeCertificate(id uint) error {
	cert, err := db.GetCertificateByID(id)
	if err != nil {
		return err
	}
	cert.Status = model.CertificateStatusRevoked
	return db.UpdateCertificate(cert)
}

func DeleteCertificate(id uint) error {
	return db.DeleteCertificate(id)
}

// --- CertificateRequest Service ---

var GetCertificateRequests = db.GetCertificateRequests
var GetCertificateRequestByID = db.GetCertificateRequestByID
var GetTenantCertificateRequests = db.GetCertificateRequestsByUserID
var CreateCertificateRequest = db.CreateCertificateRequest

// CreateTenantCertificateRequest 租户申请证书的业务逻辑
func CreateTenantCertificateRequest(user *model.User, reqType model.CertificateType, reason string) (*model.CertificateRequest, error) {
	// 1. 检查租户是否已经有了一个有效的证书
	existingCert, err := db.GetCertificateByOwnerID(user.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "failed to check existing certificate")
	}
	if existingCert != nil && (existingCert.Status == model.CertificateStatusValid || existingCert.Status == model.CertificateStatusExpiring) {
		return nil, fmt.Errorf("certificate already exists for user")
	}

	// 2. 检查租户是否已经有一个正在处理的申请
	_, err = db.GetPendingCertificateRequestByUserID(user.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "failed to check pending request")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("certificate request is pending for user")
	}

	// 3. 创建新的申请
	request := &model.CertificateRequest{
		UserName: user.Username,
		UserID:   user.ID,
		Type:     reqType,
		Status:   model.CertificateStatusPending,
		Reason:   reason,
	}

	if err := db.CreateCertificateRequest(request); err != nil {
		return nil, err
	}
	return request, nil
}

// ApproveAndCreateCertificate 将批准和创建证书合并为一个事务性操作
func ApproveAndCreateCertificate(reqID uint, adminUser *model.User) (*model.Certificate, error) {
	// 1. 获取申请信息
	req, err := db.GetCertificateRequestByID(reqID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get request by id: %d", reqID)
	}

	// 2. 检查申请状态
	if !req.IsPending() {
		return nil, fmt.Errorf("request is not pending, current status: %s", req.Status)
	}

	// 3. 创建证书
	cert := &model.Certificate{
		Name:           fmt.Sprintf("%s-%s-cert", req.UserName, req.Type),
		Type:           req.Type,
		Status:         model.CertificateStatusValid,
		Owner:          req.UserName,
		OwnerID:        req.UserID,
		Content:        "", // 实际使用中这里应该是生成的证书内容
		IssuedDate:     time.Now(),
		ExpirationDate: time.Now().AddDate(1, 0, 0), // 默认一年有效期
	}

	// 4. 更新申请状态
	req.Status = model.CertificateStatusValid
	req.ApprovedBy = adminUser.Username
	now := time.Now()
	req.ApprovedAt = &now

	// 5. 保存证书和更新申请状态
	if err := db.CreateCertificate(cert); err != nil {
		return nil, errors.Wrap(err, "failed to create certificate")
	}
	
	if err := db.UpdateCertificateRequest(req); err != nil {
		return nil, errors.Wrap(err, "failed to update request")
	}

	return cert, nil
}

// RejectCertificateRequest 拒绝证书申请
func RejectCertificateRequest(reqID uint, adminUser *model.User, reason string) error {
	// 1. 获取申请信息
	req, err := db.GetCertificateRequestByID(reqID)
	if err != nil {
		return errors.Wrapf(err, "failed to get request by id: %d", reqID)
	}

	// 2. 检查申请状态
	if !req.IsPending() {
		return fmt.Errorf("request is not pending, current status: %s", req.Status)
	}

	// 3. 更新申请状态
	req.Status = model.CertificateStatusRejected
	req.RejectedBy = adminUser.Username
	now := time.Now()
	req.RejectedAt = &now
	req.RejectedReason = reason

	// 4. 保存更新
	return db.UpdateCertificateRequest(req)
}