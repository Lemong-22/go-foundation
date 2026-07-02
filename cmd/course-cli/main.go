package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd adalah perintah paling luar/utama
var rootCmd = &cobra.Command{
	Use:   "course-cli",
	Short: "course-cli adalah tools terminal untuk manajemen course",
}

func main() {
	// Jalankan aplikasi CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Waduh error pas jalanin CLI:", err)
		os.Exit(1)
	}
}
