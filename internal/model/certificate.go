package model

import (
	"time"
	"gorm.io/gorm"
)

// CertificateType 证书类型
type CertificateType string

const (
	CertificateTypeUser CertificateType = "user" // 用户证书
	CertificateTypeNode CertificateType = "node" // 节点证书
)

// CertificateStatus 证书状态
type CertificateStatus string

const (
	CertificateStatusPending  CertificateStatus = "pending"   // 待审批
	CertificateStatusValid    CertificateStatus = "valid"     // 有效
	CertificateStatusExpiring CertificateStatus = "expiring"  // 即将过期
	CertificateStatusRevoked  CertificateStatus = "revoked"   // 已吊销
	CertificateStatusRejected CertificateStatus = "rejected"  // 已拒绝
)

// Certificate 证书实体
type Certificate struct {
	ID             uint              `json:"id" gorm:"primaryKey"`                        // unique key
	Name           string            `json:"name" gorm:"not null;index"`                 // 证书名称
	Type           CertificateType   `json:"type" gorm:"not null;index"`                 // 证书类型
	Status         CertificateStatus `json:"status" gorm:"not null;index"`               // 证书状态
	Owner          string            `json:"owner" gorm:"not null;index"`                // 证书所有者(用户名)
	OwnerID        uint              `json:"owner_id" gorm:"index"`                      // 证书所有者ID
	Content        string            `json:"content" gorm:"type:text"`                   // 证书内容(PEM格式)
	IssuedDate     time.Time         `json:"issued_date"`                                // 颁发日期
	ExpirationDate time.Time         `json:"expiration_date"`                            // 过期日期
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	DeletedAt      gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
}

// CertificateRequest 证书申请实体
type CertificateRequest struct {
	ID             uint              `json:"id" gorm:"primaryKey"`                        // unique key
	UserName       string            `json:"user_name" gorm:"not null;index"`            // 申请人用户名
	UserID         uint              `json:"user_id" gorm:"index"`                       // 申请人用户ID
	Type           CertificateType   `json:"type" gorm:"not null"`                       // 申请证书类型
	Status         CertificateStatus `json:"status" gorm:"not null;index"`               // 申请状态
	Reason         string            `json:"reason" gorm:"type:text"`                    // 申请理由
	ApprovedBy     string            `json:"approved_by,omitempty"`                      // 审批人
	ApprovedAt     *time.Time        `json:"approved_at,omitempty"`                      // 审批时间
	RejectedBy     string            `json:"rejected_by,omitempty"`                      // 拒绝人
	RejectedAt     *time.Time        `json:"rejected_at,omitempty"`                      // 拒绝时间
	RejectedReason string            `json:"rejected_reason,omitempty" gorm:"type:text"` // 拒绝理由
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	DeletedAt      gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
}

// IsValid 检查证书是否有效
func (c *Certificate) IsValid() bool {
	return c.Status == CertificateStatusValid || c.Status == CertificateStatusExpiring
}

// IsExpired 检查证书是否已过期
func (c *Certificate) IsExpired() bool {
	return c.ExpirationDate.Before(time.Now())
}

// BeforeCreate 在创建证书前确保日期字段正确处理
func (c *Certificate) BeforeCreate(tx *gorm.DB) error {
	// 确保证书日期字段被正确处理为日期（而非时间）
	c.IssuedDate = time.Date(c.IssuedDate.Year(), c.IssuedDate.Month(), c.IssuedDate.Day(), 0, 0, 0, 0, time.UTC)
	c.ExpirationDate = time.Date(c.ExpirationDate.Year(), c.ExpirationDate.Month(), c.ExpirationDate.Day(), 0, 0, 0, 0, time.UTC)
	return nil
}

// BeforeUpdate 在更新证书前确保日期字段正确处理
func (c *Certificate) BeforeUpdate(tx *gorm.DB) error {
	// 确保证书日期字段被正确处理为日期（而非时间）
	c.IssuedDate = time.Date(c.IssuedDate.Year(), c.IssuedDate.Month(), c.IssuedDate.Day(), 0, 0, 0, 0, time.UTC)
	c.ExpirationDate = time.Date(c.ExpirationDate.Year(), c.ExpirationDate.Month(), c.ExpirationDate.Day(), 0, 0, 0, 0, time.UTC)
	return nil
}

// IsPending 检查申请是否待审批
func (cr *CertificateRequest) IsPending() bool {
	return cr.Status == CertificateStatusPending
}

// IsApproved 检查申请是否已批准
func (cr *CertificateRequest) IsApproved() bool {
	return cr.Status == CertificateStatusValid
}

// IsRejected 检查申请是否已拒绝
func (cr *CertificateRequest) IsRejected() bool {
	return cr.Status == CertificateStatusRejected
}