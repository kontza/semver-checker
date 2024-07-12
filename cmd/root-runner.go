/*
Copyright © 2024 Juha Ruotsalaien <juha.ruotsalainen@iki.fi>
*/
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/api/v4/projects/%s/packages?package_name=%s",
			viper.GetString("host"),
			viper.GetString("project"),
			args[0]),
		nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create a web request due to")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("token")))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to make a web request due to")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read response body due to")
	}
	log.Info().Int("status code", resp.StatusCode).Int("body size", len(body)).Msg("Got response:")
	log.Info().Bytes("body", body).Msg("Received")
}
