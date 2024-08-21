/*
Copyright © 2024 Juha Ruotsalainen <juha.ruotsalainen@iki.fi>
*/
package cmd

import (
	"context"
	"encoding/json"
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

type FileNode struct {
	FileName     string
	DownloadPath string
}

type PackageFiles struct {
	Nodes []FileNode
}

type Package struct {
	Id           string
	PackageFiles PackageFiles
}

type GetPackageFilesResult struct {
	Package Package
}

const LATEST = "latest"
const N_A = "NOT FOUND"
const PAGE_SIZE = 2112

func downloadFile(client *graphql.Client, ctx context.Context, node Node) {
	log.Info().Interface("package", node).Msg("Downloading")
	req := graphql.NewRequest(`
			query getPackageFiles($id: PackagesPackageID!, $first: Int) {
				package(id: $id) {
					id
					packageFiles(first: $first) {
						nodes {
							fileName
							downloadPath
						}
					}
				}
			}`)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("token")))
	req.Var("id", node.Id)
	req.Var("first", PAGE_SIZE)
	var res GetPackageFilesResult
	if err := client.Run(ctx, req, &res); err != nil {
		log.Fatal().Err(err).Msg("Failed to query packages due to")
	}
	log.Debug().Interface("response", res).Msg("Received")
	if len(res.Package.PackageFiles.Nodes) < 1 {
		fmt.Println(N_A)
	}
	if n, err := json.Marshal(res.Package.PackageFiles.Nodes[0]); err != nil {
		log.Error().Err(err).Msg("Failed to marshal Node data due to")
	} else {
		fmt.Println(string(n))
	}
}

func rootRunner(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal().Msg("No package data given!")
		fmt.Println(N_A)
	}
	tokenDefined := (len(strings.TrimSpace(viper.GetString("token"))) > 0)
	log.Info().
		Str("host", viper.GetString("host")).
		Bool("token defined", tokenDefined).
		Str("project", viper.GetString("project")).Msg("Current config:")

	var packageName string
	var packageVersion string
	splitPoint := strings.LastIndex(args[0], "@")
	if splitPoint < 0 {
		packageName = args[0]
		packageVersion = LATEST
	} else {
		packageName = args[0][:splitPoint]
		packageVersion = strings.ToLower(args[0][splitPoint+1:])
	}
	log.Debug().Str("name", packageName).Str("version", packageVersion).Msg("Package")

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
	req.Var("packageName", packageName)
	req.Var("packageType", "GENERIC")
	req.Var("first", PAGE_SIZE)
	req.Var("sort", "CREATED_DESC")
	ctx := context.Background()
	var res GetPackagesResult
	if err := client.Run(ctx, req, &res); err != nil {
		log.Fatal().Err(err).Msg("Failed to query packages due to")
	}
	log.Debug().Interface("response", res).Msg("Received")

	versions := []string{}
	for _, node := range res.Project.Packages.Nodes {
		log.Debug().Interface("node", node).Msg("Found")
		// Need to prefix with 'v' to get semver.Sort to work.
		versions = append(versions, fmt.Sprintf("v%s", node.Version))
	}
	semver.Sort(versions)
	log.Debug().Strs("sorted", versions).Msg("Versions")

	if packageVersion == LATEST {
		// [1:] strips the 'v' prefix
		packageVersion = versions[len(versions)-1][1:]
	}

	var matchedNode Node
	for _, node := range res.Project.Packages.Nodes {
		if node.Version == packageVersion {
			matchedNode = node
		}
	}
	if (Node{}) == matchedNode {
		fmt.Println(N_A)
	} else {
		downloadFile(client, ctx, matchedNode)
	}
}
