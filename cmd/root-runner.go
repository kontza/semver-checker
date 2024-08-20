/*
Copyright © 2024 Juha Ruotsalainen <juha.ruotsalainen@iki.fi>
*/
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/machinebox/graphql"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/mod/semver"
)

type Node struct {
	Id      string
	Name    string
	Version string
}

type Packages struct {
	Count int
	Nodes []Node
}

type Project struct {
	Id       string
	Packages Packages
}

type GetPackagesResult struct {
	Project Project
}

func rootRunner(cmd *cobra.Command, args []string) {
	// versions := []string{"v3.4.0", "v3.11.0", "v4.0.0", "v3.7.0", "v3.8.0", "v3.9.0", "v3.10.0", "v3.12.0"}
	// semver.Sort(versions)
	// log.Info().Strs("versions", versions).Send()
	if len(args) < 1 {
		log.Fatal().Msg("No package name given!")
	}
	tokenDefined := (len(strings.TrimSpace(viper.GetString("token"))) > 0)
	log.Info().
		Str("host", viper.GetString("host")).
		Bool("token defined", tokenDefined).
		Int("project", viper.GetInt("project")).Msg("Current config:")

	client := graphql.NewClient(fmt.Sprintf("%s/api/graphql", viper.GetString("host")))
	req := graphql.NewRequest(`
		query getPackages($fullPath: ID!, $packageName: String, $packageType: PackageTypeEnum, $first: Int, $sort: PackageSort) {
			project(fullPath: $fullPath) {
				id
				packages( packageName: $packageName packageType: $packageType first: $first sort: $sort) {
					count
					nodes {
						id
						name
						version
					}
				}
			}
		}`)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("token")))
	req.Var("fullPath", viper.GetString("project"))
	req.Var("packageName", args[0])
	req.Var("packageType", "GENERIC")
	req.Var("first", 2112)
	req.Var("sort", "CREATED_DESC")
	ctx := context.Background()
	var res GetPackagesResult
	if err := client.Run(ctx, req, &res); err != nil {
		log.Fatal().Err(err).Msg("Failed to query packages due to")
	}
	log.Info().Interface("result", res).Msg("Received query")
	versions := []string{}
	for _, node := range res.Project.Packages.Nodes {
		log.Info().Interface("node", node).Msg("Found")
		versions = append(versions, fmt.Sprintf("v%s", node.Version))
	}
	log.Info().Strs("before", versions).Msg("Versions")
	semver.Sort(versions)
	log.Info().Strs("after", versions).Msg("Versions")
}
