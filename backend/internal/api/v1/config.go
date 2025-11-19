/**
 * 系统配置API处理器
 * 
 * 用途：
 * - 系统参数配置管理
 * - 支持分类查询
 * - 运营参数热更新
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"suxin/internal/appctx"
	"suxin/internal/middleware"
	"suxin/internal/model"
	"suxin/internal/repository"
	"suxin/internal/pkg/response"
)

type createConfigReq struct {
	Category    string `json:"category" binding:"required"`
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Description string `json:"description"`
	ValueType   string `json:"value_type"`
}

type updateConfigReq struct {
	Value       string `json:"value" binding:"required"`
	Description string `json:"description"`
}

type batchUpdateConfigReq struct {
	Configs map[string]string `json:"configs"`
}

/**
 * RegisterConfigRoutes 注册系统配置路由
 * 
 * 路由列表：
 * - GET  /configs            查询所有配置
 * - GET  /configs/:category  按分类查询
 * - POST /configs            创建配置(超级管理员)
 * - PUT  /configs/:id        更新配置(超级管理员)
 * - DELETE /configs/:id      删除配置(超级管理员)
 */
func RegisterConfigRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	configRepo := repository.NewConfigRepository(ctx.DB)
	
	/**
	 * GET /configs - 查询所有配置(管理员可见)
	 */
	rg.GET("/configs", middleware.RequireAdmin(ctx), func(c *gin.Context) {
		category := c.Query("category")
		
		var configs []*model.SystemConfig
		var err error
		
		if category != "" {
			configs, err = configRepo.FindByCategory(category)
		} else {
			configs, err = configRepo.FindAll()
		}
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"configs": configs,
			"total":   len(configs),
		})
	})
	
	/**
	 * POST /configs - 创建配置(超级管理员)
	 */
	rg.POST("/configs", middleware.RequireSuperAdmin(ctx), func(c *gin.Context) {
		var req createConfigReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		config := &model.SystemConfig{
			Category:    req.Category,
			Key:         req.Key,
			Value:       req.Value,
			Description: req.Description,
			ValueType:   req.ValueType,
		}
		
		if config.ValueType == "" {
			config.ValueType = "string"
		}
		
		if err := configRepo.Create(config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "创建成功",
			"config":  config,
		})
	})
	
	/**
	 * PUT /configs/:id - 更新配置(超级管理员)
	 */
	rg.PUT("/configs/:id", middleware.RequireSuperAdmin(ctx), func(c *gin.Context) {
		configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
			return
		}
		
		var req updateConfigReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		// 查询配置
		var config model.SystemConfig
		if err := ctx.DB.First(&config, configID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
			return
		}
		
		// 更新值
		config.Value = req.Value
		if req.Description != "" {
			config.Description = req.Description
		}
		
		if err := configRepo.Update(&config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "更新成功",
			"config":  config,
		})
	})
	
	/**
	 * DELETE /configs/:id - 删除配置(超级管理员)
	 */
	rg.DELETE("/configs/:id", middleware.RequireSuperAdmin(ctx), func(c *gin.Context) {
		configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
			return
		}
		
		// 检查是否是系统配置
		var config model.SystemConfig
		if err := ctx.DB.First(&config, configID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
			return
		}
		
		if config.IsSystem {
			c.JSON(http.StatusForbidden, gin.H{"error": "系统配置不可删除"})
			return
		}
		
		if err := configRepo.Delete(uint(configID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "删除成功",
		})
	})
	
	/**
	 * POST /configs/batch - 批量更新配置(超级管理员)
	 * 用于前端一次性更新多个配置项
	 */
	rg.POST("/configs/batch", middleware.RequireSuperAdmin(ctx), func(c *gin.Context) {
		var req map[string]string
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "请求参数错误")
			return
		}
		
		// 批量更新配置
		updatedCount := 0
		for key, value := range req {
			// 查找配置
			var config model.SystemConfig
			if err := ctx.DB.Where("`key` = ?", key).First(&config).Error; err != nil {
				// 配置不存在，创建新配置
				config = model.SystemConfig{
					Category:    "system",
					Key:         key,
					Value:       value,
					ValueType:   "string",
					Description: key,
				}
				if err := configRepo.Create(&config); err != nil {
					continue
				}
			} else {
				// 配置存在，更新值
				config.Value = value
				if err := configRepo.Update(&config); err != nil {
					continue
				}
			}
			updatedCount++
		}
		
		response.SuccessWithMessage(c, gin.H{
			"updated": updatedCount,
		}, "配置更新成功")
	})
}
