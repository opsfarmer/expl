// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"github.com/muroon/expl/pkg/expl"
	"github.com/muroon/expl/pkg/expl/model"
	"github.com/muroon/expl/pkg/expl/view"

	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var explainTemplate string = `expl explain mode [option] [sql]

mode:
  simple: explain one sql 
  log:    explain sqls from log file
  log-db: explain sqls from database.

sql:  it's necessary in only "simple" mode.

option:
  -d --database:            string  database. used by any sqls. (using onlh simple mode. not using in table mapping mode)
  -H --host:                string  host used by any sqls.(using onlh simple mode. not using table in mapping mode)
  -u --user:                string  database user used by any sqls.(using onlh simple mode. not using table in mapping mode)
  -p --pass:                string  database password used by any sqls.(using onlh simple mode. not using table in mapping mode)
  -c, --conf:               string  use table mapping. This is efficient for switching hosts and databases by table. 
                                    "EXPL_CONF" environment variable as a default value.
                                      value:
                                        [mapping file path]: database-table mapping file path. default file is ./table_map.yaml.
                                      ex)
                                        -c $GOPATH/bin/table-mapping.yaml
  -l, --log:                string  log file path. This is used in log mode. "EXPL_LOG" environment variable as a default value.
  -f, --format:             string  sql format.
                                      value:
                                        simple (default):  simple is only sql.
                                        official:          sql is offical mysql sql.log's format.
                                        command:           change text using os command. format-cmd option required.
                                      ex)
                                        -f simple
                                        -f official
                                        -f command
  --format-cmd:             string  os command string used only when fomat option is "command"
                                      ex)
                                        --format-cmd "cut -c 14-"
  --filter-select-type:     strings filter results by target select types.
                                      ex)
                                        --filter-select-type simple, subquery
                                        appear results with SIMPLE or SUBQUERY selected-types.
  --filter-no-select-type:  strings filter results without target select types.
  --filter-table:           strings filter results by target tables.
                                      ex)
                                        --filter-table user, group
                                        appear results, table of which is "user" or "group".
  --filter-no-table:        strings filter results without target tables.
  --filter-type:            strings filter results by target types.
                                      ex)
                                        --filter-type index, ref
                                        appear results wich "index" or "ref" types.
  --filter-no-type:         strings filter results without target types.
  --filter-possible-keys:   strings filter results by target possible keys.
  --filter-no-possible-keys:strings filter results without target possible keys.
  --filter-key:             strings filter results by target key.
  --filter-no-key:          strings filter results without target key.
  --filter-extra:           strings filter results by target taypes.
                                      ex)
                                        --filter-extra filesort, "using where" 
                                        appear results wich "filesort" or "using where" types.
  --filter-no-extra:        strings filter results without target types.

  -U, --use-table-map:              use table-database mapping file with explain sql.
  -P, --update-table-map:           update table-database mapping file before do explain sql. use current database environment.
  -I, --ignore-error:               ignore parse sql error or explain sql error.
  -C, --combine-sql:                This is useful in log or log-db module. combine identical SQL into one.

  --option-file:                    you can use option setting file. with this file you do not have to enter the above optional parameters.
                                    "EXPL_OPTION" environment variable as a default value.
  -v, --verbose:					verbose output.
  -h, --help:                       help. show usage.
`

func validateArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("invalit parameter")
	}

	mode := args[0]
	switch mode {
	case "simple":
		if len(args) < 2 {
			return fmt.Errorf("sql is none")
		}
	case "log":
		return nil
	case "log-db":
		return nil
	default:
		return fmt.Errorf("invalid mode. mode mast be (simple, log, log-db)")
	}

	return nil
}

// explainCmd represents the explain command
var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {

		// validate parameter
		if err := validateArgs(args); err != nil {
			return err
		}

		// mode
		mode := args[0]

		// sql
		sql := ""
		if mode == "simple" {
			sql = args[1]
		}

		optionFilePath, err := cmd.Flags().GetString("option-file")
		if err != nil {
			return err
		}
		if optionFilePath == "" {
			optionFilePath = os.Getenv("EXPL_OPTION")
		}
		if optionFilePath != "" {
			viper.SetConfigFile(optionFilePath)

			if err := viper.ReadInConfig(); err != nil {
				return err
			}
		}

		expOpt := new(model.ExplainOption)

		expOpt.DB = viper.GetString("database")
		expOpt.DBHost = viper.GetString("host")
		expOpt.DBUser = viper.GetString("user")
		expOpt.DBPass = viper.GetString("pass")
		expOpt.Config = viper.GetString("conf")
		if expOpt.Config == "" {
			expOpt.Config = os.Getenv("EXPL_CONF")
		}

		// log file
		logPath := viper.GetString("log")
		if logPath == "" {
			logPath = os.Getenv("EXPL_LOG")
		}

		// format
		format := viper.GetString("format")
		formatCmd := viper.GetString("format-cmd")
		if formatCmd != "" {
			format = string(expl.FormatCommand)
		}

		// filter options
		fiOpt := new(model.ExplainFilter)

		fiOpt.SelectType = viper.GetStringSlice("filter-select-type")
		fiOpt.SelectTypeNot = viper.GetStringSlice("filter-no-select-type")
		fiOpt.Table = viper.GetStringSlice("filter-table")
		fiOpt.TableNot = viper.GetStringSlice("filter-no-table")
		fiOpt.Type = viper.GetStringSlice("filter-type")
		fiOpt.TypeNot = viper.GetStringSlice("filter-no-type")
		fiOpt.PossibleKeys = viper.GetStringSlice("filter-possible-keys")
		fiOpt.PossibleKeysNot = viper.GetStringSlice("filter-no-possible-keys")
		fiOpt.Key = viper.GetStringSlice("filter-key")
		fiOpt.KeyNot = viper.GetStringSlice("filter-no-key")
		fiOpt.Extra = viper.GetStringSlice("filter-extra")
		fiOpt.ExtraNot = viper.GetStringSlice("filter-no-extra")

		expOpt.UseTableMap = viper.GetBool("use-table-map")
		if mode == "simple" {
			if expOpt.DB != "" && expOpt.DBHost != "" && expOpt.DBUser != "" {
				expOpt.UseTableMap = false
			}
		} else {
			expOpt.UseTableMap = true
		}
		expOpt.UpdateTableMap = viper.GetBool("update-table-map")

		expOpt.NoError = viper.GetBool("ignore-error")
		expOpt.Uniq = viper.GetBool("combine-sql")

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		if verbose {
			view.RenderOptions(expOpt, fiOpt, logPath, format, formatCmd)
		}

		ctx := context.Background()

		switch mode {
		case "simple":
			err = func() error {
				if expOpt.UseTableMap {
					if expOpt.UpdateTableMap {
						if err = expl.ReloadAllTableInfo(ctx, expOpt.Config); err != nil {
							return err
						}
					}

					if err = expl.LoadDBInfo(ctx, expOpt.Config); err != nil {
						return err
					}

				} else {
					expl.SetDBOne(
						expOpt.DBHost,
						expOpt.DB,
						expOpt.DBUser,
						expOpt.DBPass,
					)
				}

				sql, err = expl.GetQueryByFormat(expl.FormatType(format), sql, formatCmd)
				if err != nil {
					return err
				}

				exp, err := expl.Explain(ctx, sql, expOpt, fiOpt)
				if err == nil {
					view.RenderExplain(exp, false)
				}
				return err
			}()
		case "log":
			err = func() error {
				if expOpt.UpdateTableMap {
					if err = expl.ReloadAllTableInfo(ctx, expOpt.Config); err != nil {
						return err
					}
				}

				if err = expl.LoadDBInfo(ctx, expOpt.Config); err != nil {
					return err
				}

				qCh, erCh := expl.LoadQueriesFromLogChannels(ctx, logPath, expl.FormatType(format), formatCmd)

				exCh, errCh := expl.ExplainChannels(ctx, qCh, expOpt, fiOpt)
			L:
				for {
					select {
					case exp, ok := <-exCh:
						if !ok {
							break L
						}
						view.RenderExplain(exp, expOpt.Uniq)
					case err = <-errCh:
						if err != nil {
							break L
						}
					case err = <-erCh:
						if err != nil {
							break L
						}
					}
				}
				return nil
			}()

		case "log-db":
			err = func() error {
				if expOpt.UpdateTableMap {
					if err = expl.ReloadAllTableInfo(ctx, expOpt.Config); err != nil {
						return err
					}
				}

				if err = expl.LoadDBInfo(ctx, expOpt.Config); err != nil {
					return err
				}

				qCh, erCh := expl.LoadQueriesFromDBChannels(ctx)

				exCh, errCh := expl.ExplainChannels(ctx, qCh, expOpt, fiOpt)
			L:
				for {
					select {
					case exp, ok := <-exCh:
						if !ok {
							break L
						}
						view.RenderExplain(exp, expOpt.Uniq)
					case err = <-errCh:
						if err != nil {
							break L
						}
					case err = <-erCh:
						if err != nil {
							break L
						}
					}
				}
				return nil

			}()
		}

		if err != nil {
			fmt.Println("error occured!")
			fmt.Println(expl.Message(err))
		}

		return err
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)
	explainCmd.SetUsageTemplate(explainTemplate)
	explainCmd.SetHelpTemplate(explainTemplate)

	//viper.SetConfigType("yml")

	viper.SetDefault("format", "simple")

	viper.SetDefault("filter-select-type", []string{})
	viper.SetDefault("filter-no-select-type", []string{})
	viper.SetDefault("filter-table", []string{})
	viper.SetDefault("filter-no-table", []string{})
	viper.SetDefault("filter-type", []string{})
	viper.SetDefault("filter-no-type", []string{})
	viper.SetDefault("filter-possible-keys", []string{})
	viper.SetDefault("filter-no-possible-keys", []string{})
	viper.SetDefault("filter-key", []string{})
	viper.SetDefault("filter-no-key", []string{})
	viper.SetDefault("filter-extra", []string{})
	viper.SetDefault("filter-no-extra", []string{})

	viper.SetDefault("use-table-map", true)
	viper.SetDefault("update-table-map", true)

	explainCmd.PersistentFlags().StringP("database", "d", "", "database")
	explainCmd.PersistentFlags().StringP("host", "H", "", "host")
	explainCmd.PersistentFlags().StringP("user", "u", "", "database user")
	explainCmd.PersistentFlags().StringP("pass", "p", "", "database password")
	explainCmd.PersistentFlags().StringP("conf", "c", "", "config. which includes database-table mapping file.")
	explainCmd.PersistentFlags().StringP("log", "l", "", "sql log file path.")
	explainCmd.PersistentFlags().StringP("format", "f", "simple", "format of the line.")
	explainCmd.PersistentFlags().StringP("format-cmd", "", "", "os command to update line.")
	explainCmd.PersistentFlags().StringSlice("filter-select-type", []string{}, "filter results by target select types.")
	explainCmd.PersistentFlags().StringSlice("filter-no-select-type", []string{}, "filter results without target select types.")
	explainCmd.PersistentFlags().StringSlice("filter-table", []string{}, "filter results by target tables.")
	explainCmd.PersistentFlags().StringSlice("filter-no-table", []string{}, "filter results without target tables.")
	explainCmd.PersistentFlags().StringSlice("filter-type", []string{}, "filter results by target types.")
	explainCmd.PersistentFlags().StringSlice("filter-no-type", []string{}, "filter results without target types.")
	explainCmd.PersistentFlags().StringSlice("filter-possible-keys", []string{}, "strings filter results by target possible keys.")
	explainCmd.PersistentFlags().StringSlice("filter-no-possible-keys", []string{}, "strings filter results without target possible keys.")
	explainCmd.PersistentFlags().StringSlice("filter-key", []string{}, "filter results by target keys.")
	explainCmd.PersistentFlags().StringSlice("filter-no-key", []string{}, "filter results without target keys.")
	explainCmd.PersistentFlags().StringSlice("filter-extra", []string{}, "strings filter results by target types.")
	explainCmd.PersistentFlags().StringSlice("filter-no-extra", []string{}, "strings filter results without target types.")
	explainCmd.PersistentFlags().BoolP("use-table-map", "U", true, "use table-database mapping file.")
	explainCmd.PersistentFlags().BoolP("update-table-map", "P", true, "update table-database mapping file before do explain sql. use current database environment.")
	explainCmd.PersistentFlags().BoolP("ignore-error", "I", false, "ignore sql error.")
	explainCmd.PersistentFlags().BoolP("combine-sql", "C", false, "This is useful in log or log-db module. combine identical SQL into one.")

	explainCmd.Flags().StringP("option-file", "", "", "option yaml file.")
	explainCmd.Flags().BoolP("verbose", "v", false, "verbose output.")

	_ = viper.BindPFlag("database", explainCmd.PersistentFlags().Lookup("database"))
	_ = viper.BindPFlag("database", explainCmd.PersistentFlags().ShorthandLookup("d"))
	_ = viper.BindPFlag("host", explainCmd.PersistentFlags().Lookup("host"))
	_ = viper.BindPFlag("host", explainCmd.PersistentFlags().ShorthandLookup("H"))
	_ = viper.BindPFlag("user", explainCmd.PersistentFlags().Lookup("user"))
	_ = viper.BindPFlag("user", explainCmd.PersistentFlags().ShorthandLookup("u"))
	_ = viper.BindPFlag("pass", explainCmd.PersistentFlags().Lookup("pass"))
	_ = viper.BindPFlag("pass", explainCmd.PersistentFlags().ShorthandLookup("p"))
	_ = viper.BindPFlag("conf", explainCmd.PersistentFlags().Lookup("conf"))
	_ = viper.BindPFlag("conf", explainCmd.PersistentFlags().ShorthandLookup("c"))
	_ = viper.BindPFlag("log", explainCmd.PersistentFlags().Lookup("log"))
	_ = viper.BindPFlag("log", explainCmd.PersistentFlags().ShorthandLookup("l"))
	_ = viper.BindPFlag("format", explainCmd.PersistentFlags().Lookup("format"))
	_ = viper.BindPFlag("format", explainCmd.PersistentFlags().ShorthandLookup("f"))
	_ = viper.BindPFlag("format-cmd", explainCmd.PersistentFlags().Lookup("format-cmd"))
	_ = viper.BindPFlag("filter-select-type", explainCmd.PersistentFlags().Lookup("filter-select-type"))
	_ = viper.BindPFlag("filter-no-select-type", explainCmd.PersistentFlags().Lookup("filter-no-select-type"))
	_ = viper.BindPFlag("filter-table", explainCmd.PersistentFlags().Lookup("filter-table"))
	_ = viper.BindPFlag("filter-no-table", explainCmd.PersistentFlags().Lookup("filter-no-table"))
	_ = viper.BindPFlag("filter-type", explainCmd.PersistentFlags().Lookup("filter-type"))
	_ = viper.BindPFlag("filter-no-type", explainCmd.PersistentFlags().Lookup("filter-no-type"))
	_ = viper.BindPFlag("filter-possible-keys", explainCmd.PersistentFlags().Lookup("filter-possible-keys"))
	_ = viper.BindPFlag("filter-no-possible-keys", explainCmd.PersistentFlags().Lookup("filter-no-possible-keys"))
	_ = viper.BindPFlag("filter-key", explainCmd.PersistentFlags().Lookup("filter-key"))
	_ = viper.BindPFlag("filter-no-key", explainCmd.PersistentFlags().Lookup("filter-no-key"))
	_ = viper.BindPFlag("filter-extra", explainCmd.PersistentFlags().Lookup("filter-extra"))
	_ = viper.BindPFlag("filter-no-extra", explainCmd.PersistentFlags().Lookup("filter-no-extra"))
	_ = viper.BindPFlag("use-table-map", explainCmd.PersistentFlags().Lookup("use-table-map"))
	_ = viper.BindPFlag("use-table-map", explainCmd.PersistentFlags().ShorthandLookup("U"))
	_ = viper.BindPFlag("update-table-map", explainCmd.PersistentFlags().Lookup("update-table-map"))
	_ = viper.BindPFlag("update-table-map", explainCmd.PersistentFlags().ShorthandLookup("P"))
	_ = viper.BindPFlag("ignore-error", explainCmd.PersistentFlags().Lookup("ignore-error"))
	_ = viper.BindPFlag("ignore-error", explainCmd.PersistentFlags().ShorthandLookup("I"))
	_ = viper.BindPFlag("combine-sql", explainCmd.PersistentFlags().Lookup("combine-sql"))
	_ = viper.BindPFlag("combine-sql", explainCmd.PersistentFlags().ShorthandLookup("C"))
}
