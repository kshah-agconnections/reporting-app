package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/ksfshah3/reporting-app/configs"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

var (
	logger, _ = zap.NewProduction()
	sugar     = logger.Sugar()
)

func main() {
	configs.SetConfig()
	sugar.Infof("starting reporting app server......")
	defer logger.Sync() // flushes buffer, if any

	router := fasthttprouter.New()
	h := &fasthttp.Server{
		Handler:            router.Handler,
		MaxRequestBodySize: 2000 * 1024 * 1024 * 1024,
	}

	router.POST(configs.Configurations.AddResultsXMLPath+"/user=:userName/password=:password", addReportingResultsXML)

	if err := h.ListenAndServe(":" + configs.Configurations.RunAppOnPort); err != nil {
		log.Panicf("error in ListenAndServe: %s", err)
	}
}

func addReportingResultsXML(ctx *fasthttp.RequestCtx) {
	sugar.Infof("creating xml allure report......")
	ctx.Response.Header.Set("Content-Type", "application/json")
	fmt.Println(string(ctx.Path()))

	userName := ctx.UserValue("userName")
	password := ctx.UserValue("password")
	if userName == configs.Configurations.AppUsername && password == configs.Configurations.AppPassword {
		ctx.Response.SetStatusCode(201)
		fh, err := ctx.FormFile("file")

		if err != nil {
			sugar.Error(err)
			ctx.Response.SetStatusCode(500)
			successResponse := "{\"success\":false,\"response\":\"File key mentioned in request body is wrong\"}"
			ctx.Write([]byte(successResponse))
		}
		if err := fasthttp.SaveMultipartFile(fh, "uploads/latestreport.tar.gz"); err != nil {
			sugar.Error(err)
			ctx.Response.SetStatusCode(500)
			successResponse := "{\"success\":false,\"response\":\"Unable to save request body file\"}"
			ctx.Write([]byte(successResponse))
		}

		out, err := exec.Command("tar", "-xzvf", "uploads/latestreport.tar.gz", "-C", ".").Output()
		if err != nil {
			sugar.Error(err)
			ctx.Response.SetStatusCode(500)
			successResponse := "{\"success\":false,\"response\":\"Unable to unzip uploaded file\"}"
			ctx.Write([]byte(successResponse))
			return
		} else {
			sugar.Info("Success! created zip file into uploads folder")
			sugar.Info(string(out))
		}

		out, err = exec.Command("cp", "-r", "allure-report/history", "allure-results").Output()
		if err != nil {
			sugar.Error(err)
		} else {
			sugar.Info("Success! Allure history folder copied")
			sugar.Info(string(out))
		}

		out, err = exec.Command("allure", "generate", "allure-results", "--clean", "-o", "allure-report").Output()
		if err != nil {
			sugar.Error(err)
			ctx.Response.SetStatusCode(500)
			successResponse := "{\"success\":false,\"response\":\"Unable to generate new report\"}"
			ctx.Write([]byte(successResponse))
			return
		} else {
			sugar.Info("Success! Allure generate new report")
			sugar.Info(string(out))
		}
		successResponse := "{\"success\":true,\"response\":\"Added results to Allure reporter\"}"
		ctx.Write([]byte(successResponse))
	} else {
		ctx.Response.SetStatusCode(401)
		successResponse := "{\"success\":false,\"response\":\"Unauthorized\"}"
		ctx.Write([]byte(successResponse))
	}

}
