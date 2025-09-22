package handles

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/server/common"
)

// --- Admin Handlers ---

// CertificateList 获取证书列表，增加了分页功能，与ListUsers风格统一
func CertificateList(c *gin.Context) {
	var req model.PageReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Validate()
	certs, total, err := op.GetCertificates(req.Page, req.PerPage)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: certs,
		Total:   total,
	})
}

// CreateCertificate 创建证书
func CreateCertificate(c *gin.Context) {
	var req struct {
		Name           string    `json:"name" binding:"required"`
		Type           string    `json:"type" binding:"required"`
		Owner          string    `json:"owner"`
		OwnerID        uint      `json:"owner_id"`
		Content        string    `json:"content" binding:"required"`
		IssuedDate     time.Time `json:"issued_date"`
		ExpirationDate time.Time `json:"expiration_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	cert := &model.Certificate{
		Name:           req.Name,
		Type:           model.CertificateType(req.Type),
		Owner:          req.Owner,
		OwnerID:        req.OwnerID,
		Content:        req.Content,
		IssuedDate:     req.IssuedDate,
		ExpirationDate: req.ExpirationDate,
		Status:         model.CertificateStatusValid,
	}

	// 调用服务层创建证书
	err := op.CreateCertificate(cert)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, cert)
}

// UpdateCertificate 更新证书
func UpdateCertificate(c *gin.Context) {
	var req struct {
		Name           string    `json:"name"`
		ExpirationDate time.Time `json:"expiration_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	cert, err := op.GetCertificateByID(uint(id))
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	cert.Name = req.Name
	cert.ExpirationDate = req.ExpirationDate
	err = op.UpdateCertificate(cert)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, cert)
}

// DeleteCertificate 删除证书
func DeleteCertificate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	err = op.DeleteCertificate(uint(id))
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

// RevokeCertificate 吊销证书
func RevokeCertificate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	err = op.RevokeCertificate(uint(id))
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

// CertificateRequestList 获取证书申请列表
func CertificateRequestList(c *gin.Context) {
	var req model.PageReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Validate()
	requests, total, err := op.GetCertificateRequests(req.Page, req.PerPage)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: requests,
		Total:   total,
	})
}

// CreateCertificateRequest 创建证书申请
func CreateCertificateRequest(c *gin.Context) {
	var req struct {
		UserName string                `json:"user_name" binding:"required"`
		UserID   uint                  `json:"user_id"`
		Type     model.CertificateType `json:"type" binding:"required"`
		Reason   string                `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	request := &model.CertificateRequest{
		UserName: req.UserName,
		UserID:   req.UserID,
		Type:     req.Type,
		Reason:   req.Reason,
		Status:   model.CertificateStatusPending,
	}

	// 调用服务层创建证书申请
	err := op.CreateCertificateRequest(request)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, request)
}

// ApproveCertificateRequest 批准证书申请
func ApproveCertificateRequest(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	// 使用与项目其他部分一致的方式获取用户上下文
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	
	_, err = op.ApproveAndCreateCertificate(uint(id), user)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

// RejectCertificateRequest 拒绝证书申请
func RejectCertificateRequest(c *gin.Context) {
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	// 使用与项目其他部分一致的方式获取用户上下文
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	
	err = op.RejectCertificateRequest(uint(id), user, req.Reason)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

// DownloadCertificate 下载证书
func DownloadCertificate(c *gin.Context) {
	// 这里应该实现证书下载逻辑
	// 为了简化示例，我们返回一个模拟的证书内容
	c.String(http.StatusOK, "-----BEGIN CERTIFICATE-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...\n-----END CERTIFICATE-----")
}

// --- Tenant Handlers ---

// CreateTenantCertificateRequest 租户申请证书
func CreateTenantCertificateRequest(c *gin.Context) {
	var req struct {
		Type   model.CertificateType `json:"type" binding:"required"`
		Reason string                `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	// 使用与项目其他部分一致的方式获取用户上下文
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	
	request, err := op.CreateTenantCertificateRequest(user, req.Type, req.Reason)
	if err != nil {
		// 检查特定的错误类型
		if err.Error() == "certificate already exists for user" || err.Error() == "certificate request is pending for user" {
			common.ErrorResp(c, err, 400)
			return
		}
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, request)
}

// GetTenantCertificate 获取租户证书
func GetTenantCertificate(c *gin.Context) {
	// 使用与项目其他部分一致的方式获取用户上下文
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	
	cert, err := op.GetCertificateForTenant(user.ID)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, cert)
}

// GetTenantCertificateRequests 获取租户证书申请记录
func GetTenantCertificateRequests(c *gin.Context) {
	// 使用与项目其他部分一致的方式获取用户上下文
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	
	requests, err := op.GetTenantCertificateRequests(user.ID)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, requests)
}