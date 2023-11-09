// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"net/http"

	"github.com/minio/minio/internal/grid"
	"github.com/minio/mux"
)

// Composed function registering routers for only distributed Erasure setup.
// 为分布式擦除设置注册路由器的组合函数。
func registerDistErasureRouters(router *mux.Router, endpointServerPools EndpointServerPools) {
	// Register storage REST router only if its a distributed setup.
    // 仅当其是分布式设置时才注册存储 REST 路由器。
	registerStorageRESTHandlers(router, endpointServerPools, globalGrid.Load())

	// Register peer REST router only if its a distributed setup.
    // 仅当其是分布式设置时才注册对等 REST 路由器。
	registerPeerRESTHandlers(router)

	// Register peer S3 router only if its a distributed setup.
    // 仅在分布式设置时才注册对等 S3 路由器。
	registerPeerS3Handlers(router)

	// Register bootstrap REST router for distributed setups.
    // 注册引导 REST 路由器以进行分布式设置。
	registerBootstrapRESTHandlers(router)

	// Register distributed namespace lock routers.
    // 注册分布式命名空间锁路由器。
	registerLockRESTHandlers()

	// Add grid to router
	router.Handle(grid.RoutePath, adminMiddleware(globalGrid.Load().Handler(), noGZFlag, noObjLayerFlag))
}

// List of some generic middlewares which are applied for all incoming requests.
// 全局中间件列表
var globalMiddlewares = []mux.MiddlewareFunc{
	// set x-amz-request-id header and others
    // 设置头
	addCustomHeadersMiddleware,
	// The generic tracer needs to be the first middleware to catch all requests
	// returned early by any other middleware (but after the middleware that
	// sets the amz request id).
    // http 跟踪
	httpTracerMiddleware,
	// Auth middleware verifies incoming authorization headers and routes them
	// accordingly. Client receives a HTTP error for invalid/unsupported
	// signatures.
	//
	// Validates all incoming requests to have a valid date header.
    // 验证中间件
	setAuthMiddleware,
	// Redirect some pre-defined browser request paths to a static location
	// prefix.
    // 静态重定向
	setBrowserRedirectMiddleware,
	// Adds 'crossdomain.xml' policy middleware to serve legacy flash clients.
    // 垮域处理
	setCrossDomainPolicyMiddleware,
	// Limits all body and header sizes to a maximum fixed limit
    // 限制 body 大小
	setRequestLimitMiddleware,
	// Validate all the incoming requests.
    // 验证所有请求
	setRequestValidityMiddleware,
	// Add upload forwarding middleware for site replication
    // 转发？
	setUploadForwardingMiddleware,
	// Add bucket forwarding middleware
    // 转发？
	setBucketForwardingMiddleware,
	// Add new middlewares here.
}

// configureServer handler returns final handler for the http server.
// 注册服务器服务
func configureServerHandler(endpointServerPools EndpointServerPools) (http.Handler, error) {
	// Initialize router. `SkipClean(true)` stops minio/mux from
	// normalizing URL path minio/minio#3256
	router := mux.NewRouter().SkipClean(true)
	// router := mux.NewRouter().SkipClean(true).UseEncodedPath()

	// Initialize distributed NS lock.
	if globalIsDistErasure {
		registerDistErasureRouters(router, endpointServerPools)
	}

	// Add Admin router, all APIs are enabled in server mode.
    // 添加管理路由器，所有API均在服务器模式下启用。
	registerAdminRouter(router, true)

	// Add healthCheck router
    // 添加健康检查路由器
	registerHealthCheckRouter(router)

	// Add server metrics router
    // 添加服务器指标路由器
	registerMetricsRouter(router)

	// Add STS router always.
    // 始终添加 STS(Security Token Service) 路由器。
	registerSTSRouter(router)

	// Add KMS router
	registerKMSRouter(router)

	// Add API router
    // s3 协议 api 注册
	registerAPIRouter(router)

    // 注册所有插件
	router.Use(globalMiddlewares...)

	return router, nil
}
