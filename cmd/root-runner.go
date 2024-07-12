/*
Copyright © 2024 Juha Ruotsalaien <juha.ruotsalainen@iki.fi>
*/
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

func rootRunner(cmd *cobra.Command, args []string) {
	versions := []string{"v3.4.0", "v3.11.0", "v4.0.0", "v3.7.0", "v3.8.0", "v3.9.0", "v3.10.0", "v3.12.0"}
	semver.Sort(versions)
	log.Info().Strs("versions", versions).Send()
}
