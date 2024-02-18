package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func cmdRead() *cobra.Command {
	var srcpath string

	var cmdRead = &cobra.Command{
		Use:   "read -p [resource path]",
		Short: "Reads the version from resource file",
		Run: func(cmd *cobra.Command, args []string) {
			if srcpath == "" {
				fmt.Println("The resource file path is missing")
				return
			}

			ver := readVersion(srcpath)
			if ver == "" {
				fmt.Printf("No version string was found in the file %s\n", srcpath)
			} else {
				fmt.Println(ver)
			}
		},
	}

	cmdRead.Flags().StringVarP(&srcpath, "path", "p", "", "The resource file path")

	return cmdRead
}

func cmdChange() *cobra.Command {

	var srcpath, rcversion string

	var cmdWrite = &cobra.Command{
		Use:   "change -p [resource path] -v [version]",
		Short: "Change the version in the resource file",
		Run: func(cmd *cobra.Command, args []string) {
			if srcpath == "" {
				fmt.Println("The resource path is missing")
				return
			}

			if rcversion == "" {
				fmt.Println("The version parameter is missing")
				return
			}

			start := time.Now()
			count := changeVersion(srcpath, rcversion)
			if count > 0 {
				fmt.Printf("%d Updated file(s) in %s\n", count, time.Since(start))
			} else {
				fmt.Println("No file was updated")
			}
		},
	}

	cmdWrite.Flags().StringVarP(&srcpath, "path", "p", "", "The resource path")
	cmdWrite.Flags().StringVarP(&rcversion, "version", "v", "", "The version to be updated")

	return cmdWrite
}

func main() {

	var cmdRoot = &cobra.Command{
		Use:  "rcversion [command] [-p filepath/directory] [-v version]",
		Long: "Read and change the version in MSVC resouce file (.rc).\n*The file must be enconded in ANSI or UTF-8",
	}

	cmdRoot.AddCommand(cmdRead())
	cmdRoot.AddCommand(cmdChange())

	cmdRoot.Execute()
}
