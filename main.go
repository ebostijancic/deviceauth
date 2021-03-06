// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package main

import (
	"context"
	"flag"

	"github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/deviceauth/config"
	"github.com/mendersoftware/deviceauth/store/mongo"
)

func main() {
	var configPath string
	var printVersion bool
	var devSetup bool
	var debug bool

	flag.StringVar(&configPath, "config",
		"",
		"Configuration file path. Supports JSON, TOML, YAML and HCL formatted configs.")
	flag.BoolVar(&printVersion, "version",
		false, "Show version")
	flag.BoolVar(&devSetup, "dev",
		false, "Use development setup")
	flag.BoolVar(&debug, "debug",
		false, "Enable debug logging")

	flag.Parse()

	log.Setup(debug)

	l := log.New(log.Ctx{})

	ctx := context.Background()

	HandleConfigFile(configPath, devSetup, l)

	l.Printf("Device Authentication Service, version %s starting up",
		CreateVersionString())

	db, err := mongo.NewDataStoreMongo(config.Config.GetString(SettingDb))
	if err != nil {
		l.Fatal("failed to connect to db")
	}

	err = db.Migrate(ctx, mongo.DbVersion)
	if err != nil {
		l.Fatalf("failed to run migrations: %v", err)
	}

	l.Fatal(RunServer(config.Config))
}

func HandleConfigFile(configPath string, devSetup bool, l *log.Logger) {
	err := config.FromConfigFile(configPath, configDefaults)
	if err != nil {
		l.Fatalf("error loading configuration: %s", err)
	}

	if devSetup == true {
		l.Infof("setting up development configuration")
		config.Config.Set(SettingMiddleware, EnvDev)
	}

	// Enable setting conig values by environment variables
	config.Config.SetEnvPrefix("DEVICEAUTH")
	config.Config.AutomaticEnv()
}
