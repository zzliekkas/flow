package app

import (
	"net/http"

	"github.com/zzliekkas/flow"
	"github.com/zzliekkas/flow/cloud/providers"
)

// Kd100Controller 示例控制器
type Kd100Controller struct {
	Kd100 *providers.Kd100Service // 依赖注入
}

// RegisterRoutes 注册路由
func (c *Kd100Controller) RegisterRoutes(router flow.RouterGroup) {
	router.POST("/express/track", c.Track)
}

// Track 查询物流轨迹接口
func (c *Kd100Controller) Track(ctx *flow.Context) {
	var req struct {
		Com string `json:"com" binding:"required"`
		Num string `json:"num" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, flow.H{"error": err.Error()})
		return
	}
	result, err := c.Kd100.QueryTrack(req.Com, req.Num)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, flow.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, flow.H{"data": result})
}
