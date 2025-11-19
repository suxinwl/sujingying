/**
 * 邀请码API处理器
 * 
 * 用途：
 * - 邀请码管理
 * - 邀请关系查询
 * - 团队层级查看
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package v1

import (
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"

	"suxin/internal/appctx"
	"suxin/internal/service"
)

/**
 * RegisterInvitationRoutes 注册邀请码路由
 * 
 * 路由列表：
 * - GET  /invitation/my-code        获取我的邀请码
 * - GET  /invitation/my-invitees    获取我邀请的人
 * - GET  /invitation/team-info      获取团队信息
 * - GET  /invitation/team-members   获取直接下级
 */
func RegisterInvitationRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	invitationSvc := service.NewInvitationService(ctx)
	
	/**
	 * GET /invitation/my-code - 获取我的邀请码
	 * 
	 * 响应：
	 * {
	 *   "code": "A1B2C3D4",
	 *   "invite_count": 10,
	 *   "register_count": 10,
	 *   "share_url": "https://app.com/register?code=A1B2C3D4"
	 * }
	 */
	rg.GET("/invitation/my-code", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		
		invCode, err := invitationSvc.GetMyInvitationCode(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// 构建分享URL（可配置）
		shareURL := "https://app.suxinying.com/register?code=" + invCode.Code
		// 生成二维码（PNG）并转为 data URL，便于前端直接展示
		var qrDataURL string
		if png, err := qrcode.Encode(shareURL, qrcode.Medium, 256); err == nil {
			b64 := base64.StdEncoding.EncodeToString(png)
			qrDataURL = "data:image/png;base64," + b64
		}
		
		c.JSON(http.StatusOK, gin.H{
			"code":           invCode.Code,
			"invite_count":   invCode.InviteCount,
			"register_count": invCode.RegisterCount,
			"share_url":      shareURL,
			"qr_code":        qrDataURL,
		})
	})
	
	/**
	 * GET /invitation/my-invitees - 获取我邀请的人
	 * 
	 * 查询参数：
	 * - limit: 每页数量（默认20）
	 * - offset: 偏移量（默认0）
	 * 
	 * 响应：
	 * {
	 *   "invitees": [
	 *     {
	 *       "id": 1,
	 *       "invitee_id": 123,
	 *       "invite_code": "A1B2C3D4",
	 *       "reward_amount": 10.00,
	 *       "reward_status": "issued",
	 *       "created_at": "2025-11-18"
	 *     }
	 *   ],
	 *   "total": 10
	 * }
	 */
	rg.GET("/invitation/my-invitees", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		invitees, err := invitationSvc.GetMyInvitees(userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"invitees": invitees,
			"total":    len(invitees),
		})
	})
	
	/**
	 * GET /invitation/team-info - 获取团队信息
	 * 
	 * 响应：
	 * {
	 *   "user_id": 1,
	 *   "parent_id": 0,
	 *   "level": 1,
	 *   "direct_count": 10,
	 *   "team_count": 50,
	 *   "team_points": 1000.50
	 * }
	 */
	rg.GET("/invitation/team-info", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		
		teamInfo, err := invitationSvc.GetTeamInfo(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"user_id":      teamInfo.UserID,
			"parent_id":    teamInfo.ParentID,
			"level":        teamInfo.Level,
			"direct_count": teamInfo.DirectCount,
			"team_count":   teamInfo.TeamCount,
			"team_points":  teamInfo.TeamPoints,
		})
	})
	
	/**
	 * GET /invitation/team-members - 获取直接下级
	 * 
	 * 响应：
	 * {
	 *   "members": [
	 *     {
	 *       "user_id": 123,
	 *       "level": 2,
	 *       "direct_count": 5,
	 *       "team_count": 20,
	 *       "created_at": "2025-11-18"
	 *     }
	 *   ],
	 *   "total": 10
	 * }
	 */
	rg.GET("/invitation/team-members", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		
		members, err := invitationSvc.GetDirectTeamMembers(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"members": members,
			"total":   len(members),
		})
	})
}
