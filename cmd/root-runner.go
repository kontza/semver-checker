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
	DownloadPath string
	FileName     string
	Url          string
	Version      string
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
const N_A = "[]"
const PAGE_SIZE = 2112

func getPackageFiles(client *graphql.Client, ctx context.Context, node Node) {
	log.Info().Interface("package", node).Msg("Get files for")
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", viper.GetString(TOKEN)))
	req.Var("id", node.Id)
	req.Var("first", PAGE_SIZE)
	var res GetPackageFilesResult
	if err := client.Run(ctx, req, &res); err != nil {
		log.Fatal().Err(err).Msg("Failed to query packages due to")
	}
	log.Debug().Interface("response", res).Msg("Received")
	results := []FileNode{}
	if len(res.Package.PackageFiles.Nodes) < 1 {
		fmt.Println(N_A)
	} else {
		for _, fileNode := range res.Package.PackageFiles.Nodes {
			fileNode.Url = fmt.Sprintf("%s/api/v4/projects/%s/packages/generic/%s/%s/%s",
				viper.GetString(HOST),
				viper.GetString(PROJECT_ID),
				node.Name,
				node.Version,
				fileNode.FileName,
			)
			fileNode.Version = node.Version
			log.Debug().Interface("fileNode", fileNode).Msg("Updated")
			results = append(results, fileNode)
		}
	}
	if n, err := json.Marshal(results); err != nil {
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

	for _, key := range []string{TOKEN, HOST, PROJECT, PROJECT_ID} {
		isDefined := (len(strings.TrimSpace(viper.GetString(key))) > 0)
		if !isDefined {
			log.Fatal().Msgf("No %s defined in the config file, nor in a SEMCHK-prefixed environment variable! Cannot continue.", key)
		}
	}

	log.Info().
		Str(HOST, viper.GetString(HOST)).
		Bool("token defined", true).
		Str(PROJECT, viper.GetString(PROJECT)).Msg("Current config:")

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

	client := graphql.NewClient(fmt.Sprintf("%s/api/graphql", viper.GetString(HOST)))
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", viper.GetString(TOKEN)))
	req.Var("fullPath", viper.GetString(PROJECT))
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
	if len(versions) < 1 {
		fmt.Println(N_A)
		return
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
		getPackageFiles(client, ctx, matchedNode)
	}
}
