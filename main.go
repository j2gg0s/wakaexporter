package main

import "github.com/j2gg0s/wakaexporter/cmd"

func main() {
	root := cmd.NewRootCommand()

	root.AddCommand(cmd.NewSyncCommand())

	if err := root.Execute(); err != nil {
		panic(err)
	}
}
