package common

import "fmt"

const purple = "\033[0;35m"

// PrintBanner mirrors print_banner(title, subtitle) from common.sh.
func PrintBanner(title, subtitle string) {
	fmt.Print(purple)
	fmt.Println(`██████╗ ██╗      █████╗ ███████╗███████╗██╗   ██╗`)
	fmt.Println(`██╔══██╗██║     ██╔══██╗██╔════╝██╔════╝╝██╗ ██╔╝`)
	fmt.Println(`██████╔╝██║     ███████║█████╗  ███████╗ ╚████╔╝ `)
	fmt.Println(`██╔═══╝ ██║     ██╔══██║██╔══╝  ╚════██║  ╚██╔╝  `)
	fmt.Println(`██║     ███████╗██║  ██║███████╗███████║   ██║   `)
	fmt.Println(`╚═╝     ╚══════╝╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝   `)
	fmt.Println()
	if title != "" {
		fmt.Printf("🏛️  %s\n", title)
	}
	if subtitle != "" {
		fmt.Printf("   %s\n", subtitle)
	}
	fmt.Print(colorReset)
}
