package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	router.DELETE(configs.Configurations.DeleteResultsXMLPath+"/user=:userName/password=:password", deleteReportingResultsXML)

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
			successResponse := "{\"success\":false,\"response\":\"Zip file of allure results file missing in Request Body\"}"
			ctx.Write([]byte(successResponse))
			return
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
		data, err := ioutil.ReadFile("allure-report/index.html")
		if err != nil {
			sugar.Error("File not present", err)
		} else {
			indexHTML := string(data)
			indexHTML = strings.Replace(indexHTML, "Allure Report", configs.Configurations.ProjectName, 1)
			err = ioutil.WriteFile("allure-report/index.html", []byte(indexHTML), 0644)
			if err != nil {
				sugar.Error("File not present and Unable to write", err)
			} else {
				sugar.Info("Allure report Title Updated")
			}
		}
		successResponse := "{\"success\":true,\"response\":\"Added results to Allure reporter\"}"
		ctx.Write([]byte(successResponse))
	} else {
		ctx.Response.SetStatusCode(401)
		successResponse := "{\"success\":false,\"response\":\"Unauthorized\"}"
		ctx.Write([]byte(successResponse))
	}
}

func deleteReportingResultsXML(ctx *fasthttp.RequestCtx) {
	sugar.Infof("deleting xml allure report......")
	ctx.Response.Header.Set("Content-Type", "application/json")
	fmt.Println(string(ctx.Path()))

	userName := ctx.UserValue("userName")
	password := ctx.UserValue("password")
	if userName == configs.Configurations.AppUsername && password == configs.Configurations.AppPassword {
		ctx.Response.SetStatusCode(201)
		RemoveContents("allure-report")
		RemoveContents("allure-results")
		successResponse := "{\"success\":true,\"response\":\"Existing Allure report is flushed\"}"
		ctx.Write([]byte(successResponse))
	} else {
		ctx.Response.SetStatusCode(401)
		successResponse := "{\"success\":false,\"response\":\"Unauthorized\"}"
		ctx.Write([]byte(successResponse))
	}
}

func RemoveContents(dir string) {
	d, err := os.Open(dir)
	if err != nil {
		sugar.Error("Unable to find directory")
		return
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		sugar.Error("Unable to find file names")
		return
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			sugar.Error("Unable to Delete files")
			return
		}
	}
	sugar.Info("Folder files deleted " + dir)
}
