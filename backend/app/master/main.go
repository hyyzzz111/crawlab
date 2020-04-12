package main

import (
	"crawlab/app/master/cmd"
	"github.com/apex/log"
	"github.com/davecgh/go-spew/spew"
	"os"
)

func main(){
	app := cmd.NewApp()
	if err := app.Run(os.Args); err != nil {
		spew.Dump(err)
		log.Fatalf("程序启动失败 error: %s",err)
	}
}