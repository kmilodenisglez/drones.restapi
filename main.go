package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/iris-contrib/swagger/v12"              // swagger middleware for Iris
	"github.com/iris-contrib/swagger/v12/swaggerFiles" // swagger embed files
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kmilodenisglez/drones.restapi/api/endpoints"
	"github.com/kmilodenisglez/drones.restapi/api/middlewares"
	_ "github.com/kmilodenisglez/drones.restapi/docs"
	"github.com/kmilodenisglez/drones.restapi/service/utils"

	_ "github.com/lib/pq"
)

// @title drones
// @version 0.1
// @description REST API that allows clients to communicate with drones (i.e. **dispatch controller**)

// @contact.name Kmilo Denis Glez
// @contact.url https://github.com/kmilodenisglez
// @contact.email kmilo.denis.glez@gmail.com

// @authorizationurl https://example.com/oauth/authorize

// TIPS This Ip here 👇🏽  must be change when compiling to deploy, can't figure out how to do it dynamically with Iris.

// @BasePath /
func main() {
	// region ======== GLOBALS ===============================================================
	v := validator.New() // Validator instance. Reference https://github.com/kataras/iris/wiki/Model-validation | https://github.com/go-playground/validator

	app := iris.New() // App instance
	app.Validator = v // Register validation on the iris app

	// Services
	svcConfig := utils.NewSvcConfig()              // Creating Configuration Service
	svcResponse := utils.NewSvcResponse(svcConfig) // Creating Response Service
	// endregion =============================================================================

	// region ======== MIDDLEWARES ===========================================================
	// Our custom CORS middleware.
	crs := func(ctx iris.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if ctx.Method() == iris.MethodOptions {
			ctx.Header("Access-Control-Methods",
				"POST, PUT, PATCH, DELETE")

			ctx.Header("Access-Control-Allow-Headers",
				"Access-Control-Allow-Origin,Content-Type,authorization")

			ctx.Header("Access-Control-Max-Age",
				"86400")

			ctx.StatusCode(iris.StatusNoContent)
			return
		}

		ctx.Next()
	}

	// built-ins
	app.Use(logger.New())
	app.UseRouter(crs) // Recovery middleware recovers from any panics and writes a 500 if there was one.

	// custom middleware
	MdwAuthChecker := middlewares.NewAuthCheckerMiddleware([]byte(svcConfig.JWTSignKey))

	// endregion =============================================================================

	// region ======== ENDPOINT REGISTRATIONS ================================================

	endpoints.NewAuthHandler(app, &MdwAuthChecker, svcResponse, svcConfig)
	endpoints.NewBlockchainTxsHandler(app, &MdwAuthChecker, svcResponse, svcConfig) // Blockchain transactions handlers
	// endregion =============================================================================

	// region ======== SWAGGER REGISTRATION ==================================================
	// sc == swagger config
	sc := &swagger.Config{
		DeepLinking: true,
		URL:         "http://" + svcConfig.ApiDocIp + ":" + svcConfig.DappPort + "/swagger/apidoc.json", // The url pointing to API definition
	}

	// use swagger middleware to
	app.Get("/swagger/{any:path}", swagger.CustomWrapHandler(sc, swaggerFiles.Handler))
	// endregion =============================================================================

	addr := fmt.Sprintf("%s:%s", svcConfig.ApiDocIp, svcConfig.DappPort)

	// run localhost
	app.Listen(addr)
}
