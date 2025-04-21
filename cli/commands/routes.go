package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/zzliekkas/flow/cli"
)

// NewRoutesCommand 创建路由列表命令
func NewRoutesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "routes",
		Aliases: []string{"route", "route:list"},
		Short:   "显示所有注册的路由",
		Long:    `显示应用中所有注册的路由，包括HTTP方法、URL路径、处理器和中间件信息。`,
		Run:     listRoutes,
	}

	// 添加命令行标志
	cmd.Flags().StringP("method", "m", "", "按HTTP方法筛选 (GET, POST, PUT, DELETE等)")
	cmd.Flags().StringP("path", "p", "", "按路径筛选 (支持部分匹配)")
	cmd.Flags().BoolP("verbose", "v", false, "显示详细信息，包括中间件")
	cmd.Flags().BoolP("reverse", "r", false, "反向排序")

	return cmd
}

// 路由信息结构体
type routeInfo struct {
	Method     string
	Path       string
	Handler    string
	Middleware []string
}

// listRoutes 列出所有路由
func listRoutes(cmd *cobra.Command, args []string) {
	// 获取筛选条件
	methodFilter, _ := cmd.Flags().GetString("method")
	pathFilter, _ := cmd.Flags().GetString("path")
	verbose, _ := cmd.Flags().GetBool("verbose")
	reverse, _ := cmd.Flags().GetBool("reverse")

	// 收集路由信息
	routes := collectRoutes()

	// 筛选路由
	var filteredRoutes []routeInfo
	for _, route := range routes {
		if methodFilter != "" && !strings.EqualFold(route.Method, methodFilter) {
			continue
		}
		if pathFilter != "" && !strings.Contains(strings.ToLower(route.Path), strings.ToLower(pathFilter)) {
			continue
		}
		filteredRoutes = append(filteredRoutes, route)
	}

	// 排序
	sort.Slice(filteredRoutes, func(i, j int) bool {
		if filteredRoutes[i].Path == filteredRoutes[j].Path {
			return filteredRoutes[i].Method < filteredRoutes[j].Method
		}
		result := filteredRoutes[i].Path < filteredRoutes[j].Path
		if reverse {
			return !result
		}
		return result
	})

	// 打印路由
	if len(filteredRoutes) == 0 {
		cli.PrintInfo("没有找到匹配的路由")
		return
	}

	cli.PrintSuccess("找到 %d 个路由", len(filteredRoutes))

	// 使用tabwriter格式化输出
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "METHOD\tPATH\tHANDLER")
	fmt.Fprintln(w, "------\t----\t-------")

	for _, route := range filteredRoutes {
		fmt.Fprintf(w, "%s\t%s\t%s\n", route.Method, route.Path, route.Handler)

		// 显示中间件（如果启用详细模式）
		if verbose && len(route.Middleware) > 0 {
			fmt.Fprintf(w, "\t└── 中间件: %s\n", strings.Join(route.Middleware, ", "))
		}
	}
	w.Flush()
}

// collectRoutes 收集应用中所有路由信息
// 注意：这是一个框架占位函数，实际实现需要根据框架路由系统来完成
func collectRoutes() []routeInfo {
	// 在实际实现中，这里应该连接到框架的路由收集器
	// 目前返回一些示例数据用于演示
	return []routeInfo{
		{Method: "GET", Path: "/", Handler: "HomeController@index", Middleware: []string{"web", "auth"}},
		{Method: "GET", Path: "/users", Handler: "UserController@index", Middleware: []string{"web", "auth"}},
		{Method: "POST", Path: "/users", Handler: "UserController@store", Middleware: []string{"web", "auth", "csrf"}},
		{Method: "GET", Path: "/users/:id", Handler: "UserController@show", Middleware: []string{"web", "auth"}},
		{Method: "PUT", Path: "/users/:id", Handler: "UserController@update", Middleware: []string{"web", "auth", "csrf"}},
		{Method: "DELETE", Path: "/users/:id", Handler: "UserController@destroy", Middleware: []string{"web", "auth", "csrf"}},
		{Method: "GET", Path: "/api/users", Handler: "Api\\UserController@index", Middleware: []string{"api", "auth:api"}},
		{Method: "POST", Path: "/api/users", Handler: "Api\\UserController@store", Middleware: []string{"api", "auth:api"}},
	}
}
